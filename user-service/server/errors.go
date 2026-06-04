package server

import (
	"errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"ourvideos/user-service/service"
)

func toGRPCError(err error) error {
	switch {
	case errors.Is(err, service.ErrUserNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, service.ErrUserAlreadyExists):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, service.ErrUserInvalidPassword):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, service.ErrorUpdateFailed):
		return status.Error(codes.FailedPrecondition, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}
