package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"log"
	"net/http"
	"ourvideos/proto/danmaku"
	"strconv"
	"sync"
)

// ============================================================
// DanmakuHub — WebSocket 弹幕中心
// ============================================================
// 房间模型：一个视频 = 一个 room，key = "video:57"
// 每个 room 维护当前在线的 WebSocket 连接列表
// 消息通过 Redis Pub/Sub 实现跨进程广播（单机可以用 map，但 Pub/Sub 扩展性好）
// ============================================================

type DanmakuHub struct {
	rooms    map[string]map[*websocket.Conn]bool
	mu       sync.RWMutex
	RDB      *redis.Client
	upgrader websocket.Upgrader
}

func NewDanmakuHub(rdb *redis.Client) *DanmakuHub {
	return &DanmakuHub{
		rooms: make(map[string]map[*websocket.Conn]bool),
		RDB:   rdb,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true }, // 允许所有来源
		},
	}
}

// ============================================================
// HandleWS — WebSocket 入口
// ============================================================
// 路由：GET /ws/danmaku/:id
// 一个连接 = 两个 goroutine：
//   - readPump：读浏览器发的弹幕 → 存 Redis + 广播
//   - writePump：读 Redis Pub/Sub → 推给浏览器
//
// ============================================================
func (h *DanmakuHub) HandleWS(c *gin.Context) {
	fmt.Println("路由调用")
	videoID := c.Param("id")

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("[WS] 升级失败: %v", err)
		return
	}
	defer conn.Close()

	roomKey := "danmaku:room:" + videoID
	ctx := context.Background()

	//加入房间
	h.JoinRoom(roomKey, conn)
	defer h.LeaveRoom(roomKey, conn)

	//推送历史弹幕
	history, _ := h.RDB.LRange(ctx, "danmaku:history:"+videoID, -50, -1).Result()
	for _, msg := range history {
		if err := conn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
			break
		}
	}

	//订阅redis频道
	channel := "danmaku:" + videoID
	pubsub := h.RDB.Subscribe(ctx, channel)
	defer pubsub.Close()

	//双goroutine
	done := make(chan struct{})

	// 读浏览器 → 发到 Redis
	go func() {
		defer close(done)
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				log.Printf("[Danmaku] 读取消息失败，连接断开: %v", err)
				break
			}
			log.Printf("[Danmaku] 收到浏览器消息: %s", string(msg))

			// 解析 JSON 提取必要字段
			var dm danmaku.Danmaku
			if err := json.Unmarshal(msg, &dm); err != nil {
				log.Printf("[Danmaku] JSON解析失败: %v, 原始内容: %s", err, string(msg))
				continue
			}
			dm.VideoId, _ = strconv.ParseUint(videoID, 10, 64)

			// 存历史（最多 500 条）
			key := "danmaku:history:" + videoID
			h.RDB.LPush(ctx, key, string(msg))
			h.RDB.LTrim(ctx, key, 0, 499)

			// 广播到 Redis Pub/Sub
			res := h.RDB.Publish(ctx, channel, string(msg))
			log.Printf("[Danmaku] Publish结果: err=%v subscribers=%d", res.Err(), res.Val())
		}
	}()

	// 读 Redis → 发给浏览器
	ch := pubsub.Channel()
	log.Printf("[Danmaku] 等待PubSub消息 channel=%s", channel)
	for {
		select {
		case msg, ok := <-ch:
			if !ok {
				log.Printf("[Danmaku] PubSub频道关闭")
				return
			}
			log.Printf("[Danmaku] PubSub收到, 推给浏览器: %s", msg.Payload)
			if err := conn.WriteMessage(websocket.TextMessage, []byte(msg.Payload)); err != nil {
				log.Printf("[Danmaku] 推浏览器失败: %v", err)
				return
			}
		case <-done:
			log.Printf("[Danmaku] readPump退出，writePump关闭")
			return
		}
	}

}

// ============================================================
// 房间管理
// ============================================================
func (h *DanmakuHub) JoinRoom(roomKey string, conn *websocket.Conn) {
	h.mu.Lock()
	if h.rooms[roomKey] == nil {
		h.rooms[roomKey] = make(map[*websocket.Conn]bool)
	}
	h.rooms[roomKey][conn] = true
	h.mu.Unlock()
}

func (h *DanmakuHub) LeaveRoom(roomKey string, conn *websocket.Conn) {
	h.mu.Lock()
	if h.rooms[roomKey] != nil {
		delete(h.rooms[roomKey], conn)
		if len(h.rooms[roomKey]) == 0 {
			delete(h.rooms, roomKey)
		}

	}
	h.mu.Unlock()
}
