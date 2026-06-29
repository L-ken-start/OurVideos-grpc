package main

import (
	"context"
	"github.com/spf13/viper"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"ourvideos/comment-service/model"
	"ourvideos/comment-service/repository"
	"ourvideos/comment-service/server"
	"ourvideos/comment-service/service"
	"ourvideos/proto/comment"
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
	//dsn := os.Getenv("MYSQL_COMMENT_DSN")
	//if dsn == "" {
	//	dsn = "root:123456@tcp(127.0.0.1:3306)/comment_db?charset=utf8&parseTime=True&loc=Local"
	//}
	dsn := viper.GetString("mysql.comment_dsn")
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

	lis, err := net.Listen("tcp", viper.GetString("service.comment_port"))
	if err != nil {
		log.Fatalf("监听失败 %v", err)
	}

	s := grpc.NewServer(grpc.UnaryInterceptor(grpcInterceptor))
	comment.RegisterCommentServiceServer(s, srv)

	log.Println("Comment service listen on :50053")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}

}
