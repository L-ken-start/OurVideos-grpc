package service

import "errors"

var (
	ErrVideoNotFound = errors.New("video not found")
	ErrInvalidParam  = errors.New("invalid params")
	ErrInternal      = errors.New("internal error")
	ErrLike          = errors.New("like failed")
)
