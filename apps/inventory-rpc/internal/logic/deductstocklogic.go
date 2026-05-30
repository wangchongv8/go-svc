package logic

import (
	"context"

	"go-svc/apps/inventory-rpc/internal/svc"
	"go-svc/apps/inventory-rpc/inventory"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeductStockLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeductStockLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeductStockLogic {
	return &DeductStockLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeductStockLogic) DeductStock(in *inventory.DeductStockRequest) (*inventory.DeductStockResponse, error) {
	inv, err := l.svcCtx.Repository.DeductStock(l.ctx, in.ProductId, in.Quantity)
	if err != nil {
		return nil, rpcError(err)
	}

	return &inventory.DeductStockResponse{
		ProductId: inv.ProductID,
		Stock:     inv.Stock,
	}, nil
}
