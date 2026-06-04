package service

import "errors"

var (
	ErrUserNotFound        = errors.New("user not found")
	ErrUserAlreadyExists   = errors.New("user already exists")
	ErrUserInvalidPassword = errors.New("invalid password")
	Errinternal            = errors.New("internal error")
	ErrorUpdateFailed      = errors.New("update failed")
)
