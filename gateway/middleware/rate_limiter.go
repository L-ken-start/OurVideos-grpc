package middleware

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"net/http"
	"strconv"
	"time"
)

type RateLimiter struct {
	RDB    *redis.Client
	Limit  int
	Window time.Duration
}

func NewRateLimiter(rdb *redis.Client, limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{RDB: rdb, Limit: limit, Window: window}
}

// Middle返回Gin中间件
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var key string
		if userIDVal, exist := c.Get("userID"); exist {
			key = "rate:user:" + strconv.Itoa(int(userIDVal.(uint)))
		} else {
			key = "rate:ip:" + c.ClientIP()
		}

		ctx := context.Background()
		count, err := rl.RDB.Incr(ctx, key).Result()
		if err != nil {
			c.Next() //redis出错时直接放行
			return
		}

		//key的过期时间
		if count == 1 {
			rl.RDB.Expire(ctx, key, rl.Window)
		}

		//超限
		if count > int64(rl.Limit) {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "请求过于频繁"})
			c.Abort()
			return
		}

		c.Next()

	}
}
