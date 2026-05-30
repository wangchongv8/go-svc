package logic

import (
	"context"

	"go-svc/apps/inventory-rpc/internal/svc"
	"go-svc/apps/inventory-rpc/inventory"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetStockLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetStockLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetStockLogic {
	return &SetStockLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SetStockLogic) SetStock(in *inventory.SetStockRequest) (*inventory.SetStockResponse, error) {
	inv, err := l.svcCtx.Repository.SetStock(l.ctx, in.ProductId, in.Stock)
	if err != nil {
		return nil, rpcError(err)
	}

	return &inventory.SetStockResponse{
		ProductId: inv.ProductID,
		Stock:     inv.Stock,
	}, nil
}
