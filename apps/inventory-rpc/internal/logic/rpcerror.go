package logic

import (
	"errors"

	"go-svc/apps/inventory-rpc/model"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func rpcError(err error) error {
	switch {
	case errors.Is(err, model.ErrInventoryNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, model.ErrStockInsufficient):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, model.ErrQuantityInvalid),
		errors.Is(err, model.ErrStockNegative):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}
