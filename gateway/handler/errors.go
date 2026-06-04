package handler

import (
	"google.golang.org/grpc/codes"
	"net/http"
)

func toHTTPStatusCode(code codes.Code) int {
	switch code {
	case codes.NotFound:
		return http.StatusNotFound
	case codes.AlreadyExists:
		return http.StatusConflict
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.InvalidArgument:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
