package status_error

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	UserNotFound  = status.Error(codes.NotFound, "user not found")
	RoomNotFound  = status.Error(codes.NotFound, "room not found")
	RoomExists    = status.Error(codes.NotFound, "room already exists")
	InvalidName   = status.Error(codes.InvalidArgument, "name should be 3-20 symbols long")
	EmptyName     = status.Error(codes.InvalidArgument, "empty name")
	NameExists    = status.Error(codes.Unavailable, "name already exists")
	NoAccess      = status.Error(codes.PermissionDenied, "only creator can invite to room")
	Private       = status.Error(codes.PermissionDenied, "room is private")
	AlreadyInRoom = status.Error(codes.PermissionDenied, "already in room")
)

func IsStatusError(err error) bool {
	_, ok := status.FromError(err)
	return ok
}
