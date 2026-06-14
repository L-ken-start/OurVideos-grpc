package server

import (
	"errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"ourvideos/video-service/service"
)

func toGRPCError(err error) error {
	switch {
	case errors.Is(err, service.ErrVideoNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, service.ErrInvalidParam):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, service.ErrInternal):
		return status.Error(codes.Internal, err.Error())
	case errors.Is(err, service.ErrLike):
		return status.Error(codes.Internal, err.Error())

	}
	return err
}
