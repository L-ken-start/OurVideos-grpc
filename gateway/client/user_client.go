package client

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"ourvideos/proto/user"
)

func NewUserClient(addr string) (user.UserServiceClient, *grpc.ClientConn, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return nil, nil, err
	}
	client := user.NewUserServiceClient(conn)
	return client, conn, err
}
