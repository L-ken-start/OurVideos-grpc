package client

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"ourvideos/proto/video"
)

// NewVideoClient 创建 video-service 的 gRPC 客户端连接
// 与 user_client.go 完全相同的模式。
// 参数 addr 格式："localhost:50052"
func NewVideoClient(addr string) (video.VideoServiceClient, *grpc.ClientConn, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}
	client := video.NewVideoServiceClient(conn)
	return client, conn, nil
}
