package logic

import (
	"errors"

	"go-svc/apps/product-rpc/model"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func rpcError(err error) error {
	switch {
	case errors.Is(err, model.ErrProductNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, model.ErrInvalidPrice),
		errors.Is(err, model.ErrInvalidStatus),
		errors.Is(err, model.ErrNameEmpty):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}
