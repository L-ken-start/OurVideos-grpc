package client

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"ourvideos/proto/comment"
)

func NewCommentClient(addr string) (comment.CommentServiceClient, *grpc.ClientConn, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return nil, nil, err
	}

	client := comment.NewCommentServiceClient(conn)
	return client, conn, err
}
