package status_error

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	NotFound        = status.Error(codes.NotFound, "user not found")
	InvalidPassword = status.Error(codes.InvalidArgument, "invalid password")
	InvalidUsername = status.Error(codes.InvalidArgument, "invalid username")
	InvalidEmail    = status.Error(codes.InvalidArgument, "invalid email")
	NameExists      = status.Error(codes.Unavailable, "username already exists")
	EmailExists     = status.Error(codes.Unavailable, "email already exists")
	NamesAreSame    = status.Error(codes.Unavailable, "new name matches the current")
)

func IsStatusError(err error) bool {
	_, ok := status.FromError(err)
	return ok
}
