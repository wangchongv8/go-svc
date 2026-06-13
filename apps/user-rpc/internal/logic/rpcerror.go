package logic

import (
	"errors"

	"go-svc/apps/user-rpc/model"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// rpcError maps domain errors to gRPC status errors.
func rpcError(err error) error {
	switch {
	case errors.Is(err, model.ErrUsernameEmpty),
		errors.Is(err, model.ErrKratosIdentityIDEmpty):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, model.ErrUsernameExists):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, model.ErrUserNotFound):
		return status.Error(codes.NotFound, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}
