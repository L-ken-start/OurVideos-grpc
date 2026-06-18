package main

import (
	"context"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"ourvideos/proto/video"
	"ourvideos/video-service/model"
	"ourvideos/video-service/repository"
	"ourvideos/video-service/server"
	"ourvideos/video-service/service"
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
	dsn := "root:123456@tcp(127.0.0.1:3306)/video_db?charset=utf8&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		panic("failed to connect database")
	}
	if err := db.AutoMigrate(&model.Video{}); err != nil {
		panic("auto migrate failed" + err.Error())
	}
	if err := db.AutoMigrate(&model.Like{}); err != nil {
		panic("auto migrate failed" + err.Error())
	}
	// -------------------- 依赖注入 --------------------
	// 顺序：db → repo → service → server
	// 上一层的输出是下一层的输入，main.go 是唯一知道全貌的地方。
	// 这个模式叫 "Composition Root"（组合根）。
	repo := repository.NewVideoRepository(db)
	svc := service.NewVideoService(repo)
	srv := &server.VideoServer{Svc: svc}

	// -------------------- gRPC 服务 --------------------
	// 端口 50052 —— user-service 占了 50051
	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("监听失败: %v", err)
	}

	s := grpc.NewServer(grpc.UnaryInterceptor(grpcInterceptor))
	video.RegisterVideoServiceServer(s, srv)

	log.Println("Video service listening on :50052")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}

}
