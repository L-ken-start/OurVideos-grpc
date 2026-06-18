package main

import (
	"context"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"ourvideos/proto/comment"
	"ourvideos/proto/user"
	"ourvideos/user-service/model"
	"ourvideos/user-service/repository"
	"ourvideos/user-service/server"
	"ourvideos/user-service/service"
)

func grpcInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[PANIC] method=%s panic=%v", info.FullMethod, r)
			err = status.Errorf(codes.Internal, "内部错误")
		}
	}()
	start := time.Now()
	resp, err = handler(ctx, req)
	if err != nil {
		log.Printf("[GRPC-ERROR] method=%s err=%v duration=%v", info.FullMethod, err, time.Since(start))
	}
	return resp, err
}

func main() {
	//链接数据库
	dsn := "root:123456@tcp(127.0.0.1:3306)/user_db?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		panic("failed to connect database")
	}

	//自动迁移
	db.AutoMigrate(&model.User{})

	//启动grpc服务
	list, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("监听失败:%v", err.Error())
	}

	// 连接 comment-service
	commentConn, err := grpc.NewClient("localhost:50053", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("connect comment-service fail: %v", err)
	}
	defer commentConn.Close()
	commentClient := comment.NewCommentServiceClient(commentConn)

	// 依赖注入：db → repo → service → server
	repo := repository.NewUserRepository(db)
	svc := service.NewUserService(repo)

	s := grpc.NewServer(grpc.UnaryInterceptor(grpcInterceptor))
	user.RegisterUserServiceServer(s, &server.UserServer{
		Svc:           svc,
		CommentClient: commentClient,
	})

	log.Println("User service listening on :50051")
	if err := s.Serve(list); err != nil {
		log.Fatalf("服务启动失败:%v", err.Error())
	}

}
