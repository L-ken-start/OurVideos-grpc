package main

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"log"
	"ourvideos/gateway/client"
	"ourvideos/gateway/handler"
	"ourvideos/gateway/middleware"
	"time"
)

func init() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("..")
	if err := viper.ReadInConfig(); err != nil {
		log.Printf("[Config] 未找到 config.yaml，使用环境变量: %v", err)
	}
	viper.AutomaticEnv() // ENV 自动覆盖 yaml

}

func main() {
	// 连接 user-service
	userClient, userConn, err := client.NewUserClient(viper.GetString("service.user_addr"))
	if err != nil {
		log.Fatalf("connect user-service fail: %v", err)
	}
	defer userConn.Close()

	// 连接 video-service
	videoClient, videoConn, err := client.NewVideoClient(viper.GetString("service.video_addr"))
	if err != nil {
		log.Fatalf("connect video-service fail: %v", err)
	}
	defer videoConn.Close()
	//连接comment-service
	commentClient, commentConn, err := client.NewCommentClient(viper.GetString("service.comment_addr"))
	if err != nil {
		log.Fatalf("connect comment-service fail: %v", err)
	}
	defer commentConn.Close()

	rdb := redis.NewClient(&redis.Options{Addr: viper.GetString("redis.addr")})
	limiter := middleware.NewRateLimiter(rdb, 100, time.Minute)
	//创建gin引擎
	r := gin.Default()
	r.Use(middleware.RequestLogger())
	r.Use(limiter.Middleware())
	r.Static("/static", "./static")
	h := &handler.UserHandler{Client: userClient}
	videoH := &handler.VideoHandler{Client: videoClient}
	commentH := &handler.CommentHandler{Client: commentClient, UserClient: userClient}
	uploadH := &handler.UploadHandler{}
	oauthH := &handler.OAuthHandler{
		UserClient: userClient, // 复用现有 gRPC 连接
		RDB:        rdb,        // 复用现有 Redis 连接
		ClientID:   "Ov23li4xI41nSVuInJWL",
		Secret:     "asdasd",
		Callback:   "http://localhost:8888/auth/github/callback",
		Frontend:   "http://localhost:5173",
	}
	danmakuHub := handler.NewDanmakuHub(rdb)

	//第三方登录路由
	r.GET("/auth/github", oauthH.GitHubAuthorize)         // 发起登录
	r.GET("/auth/github/callback", oauthH.GitHubCallback) // GitHub 回调
	//用户注册路由
	r.POST("/user/register", h.Register)
	r.POST("/user/login", h.Login)
	//视频显示路由
	r.GET("/videos", videoH.ListVideo)
	r.GET("/videos/search", videoH.SearchVideos)
	//上传视频
	// 弹幕 WebSocket（公开，游客也能看弹幕但发弹幕需登录）
	r.GET("/ws/danmaku/:id", danmakuHub.HandleWS)

	//获取评论列表（公开）
	r.GET("/videos/:id/comments", commentH.ListComments)
	r.GET("/videos/series/:series_id", videoH.ListSeriesEpisodes)

	auth := r.Group("/", middleware.JWTAuth())
	{
		auth.Use(limiter.Middleware())
		auth.POST("/upload_video", videoH.UploadVideo)

		auth.GET("/user/:id", h.GetUser)
		auth.PUT("/user/update", h.UserUpdate)
		auth.POST("/upload", uploadH.Upload)
		auth.PUT("/user/update/password", h.PswUpdate)
		// 发评论需登录
		auth.POST("/videos/:id/comments", commentH.AddComments)
		auth.POST("/upload_videos", videoH.CreateVideo)
		auth.GET("/videos/:id", videoH.GetVideo)
		auth.POST("/videos/:id/like", videoH.LikeVideo)
		//评论业务
		auth.POST("/comments/:id/like", commentH.LikeComment)
	}

	log.Println("Gin网关启动于：8888")
	if err := r.Run(":8888"); err != nil {
		log.Fatalf("run server error,err:%v", err)
	}

}
