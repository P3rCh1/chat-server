package status_error

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	NotFound           = status.Error(codes.NotFound, "user not found")
	AuthFail           = status.Error(codes.PermissionDenied, "invalid password")
	InvalidPassword    = status.Error(codes.InvalidArgument, "password should be 8-128 symbols long")
	InvalidUsernameLen = status.Error(codes.InvalidArgument, "password should be 3-20 symbols long")
	InvalidUsernameChr = status.Error(codes.InvalidArgument, `password should  contains only letters, numbers, or "_"`)
	InvalidEmail       = status.Error(codes.InvalidArgument, "invalid email")
	NameExists         = status.Error(codes.Unavailable, "username already exists")
	EmailExists        = status.Error(codes.Unavailable, "email already exists")
	NamesAreSame       = status.Error(codes.Unavailable, "new name matches the current")
	EmptyEmail         = status.Error(codes.InvalidArgument, "email is empty")
	EmptyUsername      = status.Error(codes.InvalidArgument, "username is empty")
	EmptyPassword      = status.Error(codes.InvalidArgument, "password is empty")
)

func IsStatusError(err error) bool {
	_, ok := status.FromError(err)
	return ok
}
