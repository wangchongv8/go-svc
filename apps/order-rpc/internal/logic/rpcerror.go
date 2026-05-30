package logic

import (
	"errors"

	"go-svc/apps/order-rpc/model"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func rpcError(err error) error {
	switch {
	case errors.Is(err, model.ErrOrderNotFound):
		return status.Error(codes.NotFound, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}
