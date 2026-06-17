package main

import (
	"github.com/gin-gonic/gin"
	"log"
	"ourvideos/gateway/client"
	"ourvideos/gateway/handler"
	"ourvideos/gateway/middleware"
)

func main() {
	// 连接 user-service
	userClient, userConn, err := client.NewUserClient("localhost:50051")
	if err != nil {
		log.Fatalf("connect user-service fail: %v", err)
	}
	defer userConn.Close()

	// 连接 video-service
	videoClient, videoConn, err := client.NewVideoClient("localhost:50052")
	if err != nil {
		log.Fatalf("connect video-service fail: %v", err)
	}
	defer videoConn.Close()
	//连接comment-service
	commentClient, commentConn, err := client.NewCommentClient("localhost:50053")
	if err != nil {
		log.Fatalf("connect comment-service fail: %v", err)
	}
	defer commentConn.Close()

	//创建gin引擎
	r := gin.Default()
	r.Static("/static", "./static")
	h := &handler.UserHandler{Client: userClient}
	videoH := &handler.VideoHandler{Client: videoClient}
	commentH := &handler.CommentHandler{Client: commentClient, UserClient: userClient}
	uploadH := &handler.UploadHandler{}

	//用户注册路由
	r.POST("/user/register", h.Register)
	r.POST("/user/login", h.Login)
	//视频显示路由
	r.GET("/videos", videoH.ListVideo)
	r.GET("/videos/search", videoH.SearchVideos)
	//上传视频

	//获取评论列表（公开）
	r.GET("/videos/:id/comments", commentH.ListComments)
	r.GET("/videos/series/:series_id", videoH.ListSeriesEpisodes)

	auth := r.Group("/", middleware.JWTAuth())
	{
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
