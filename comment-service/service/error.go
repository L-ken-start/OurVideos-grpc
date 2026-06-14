package service

import "errors"

var (
	CommentNotFound = errors.New("comment not found")
	UpdateFailed    = errors.New("update failed")
	InternalError   = errors.New("internal error")
	LikeFailed      = errors.New("like failed")
)
