package main

import (
	"google.golang.org/grpc"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"net"
	"ourvideos/comment-service/model"
	"ourvideos/comment-service/repository"
	"ourvideos/comment-service/server"
	"ourvideos/comment-service/service"
	"ourvideos/proto/comment"
)

func main() {
	dsn := "root:123456@tcp(127.0.0.1:3306)/comment_db?charset=utf8&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		panic("failed to connect database")
	}
	if err := db.AutoMigrate(&model.Comment{}); err != nil {
		panic("failed to auto migrate database")
	}

	//依赖注入
	repo := repository.NewCommentRepository(db)
	svc := service.NewCommentService(repo)
	srv := &server.CommentServer{Svc: svc}

	lis, err := net.Listen("tcp", ":50053")
	if err != nil {
		log.Fatalf("监听失败 %v", err)
	}

	s := grpc.NewServer()
	comment.RegisterCommentServiceServer(s, srv)

	log.Println("Comment service listen on :50053")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}

}
