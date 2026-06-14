package server

import (
	"errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"ourvideos/comment-service/service"
)

func ErrorHandler(err error) error {
	switch {

	case errors.Is(err, service.CommentNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, service.UpdateFailed):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, service.LikeFailed):
		return status.Error(codes.FailedPrecondition, err.Error())
	default:
		return status.Error(codes.NotFound, err.Error())
	}
}
