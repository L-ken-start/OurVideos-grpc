package main

import (
	"github.com/gin-gonic/gin"
	"log"
	"ourvideos/gateway/client"
	"ourvideos/gateway/handler"
	"ourvideos/gateway/middleware"
)

func main() {
	//连接用户grpc服务
	userClient, conn, err := client.NewUserClient("localhost:50051")
	if err != nil {
		log.Fatalf("connect user fail,err:%v", err)
	}
	defer conn.Close()

	//创建gin引擎
	r := gin.Default()
	r.Static("/static", "./static")
	h := &handler.UserHandler{Client: userClient}
	uploadH := &handler.UploadHandler{}

	//注册路由
	r.POST("/user/register", h.Register)
	r.POST("/user/login", h.Login)

	auth := r.Group("/", middleware.JWTAuth())
	{
		auth.GET("/user/:id", h.GetUser)
		auth.PUT("/user/update", h.UserUpdate)
		auth.POST("/upload", uploadH.Upload)
		auth.PUT("/user/update/password", h.PswUpdate)
	}

	log.Println("Gin网关启动于：8888")
	if err := r.Run(":8888"); err != nil {
		log.Fatalf("run server error,err:%v", err)
	}

}
