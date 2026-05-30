package logic

import (
	"context"

	"go-svc/apps/inventory-rpc/internal/svc"
	"go-svc/apps/inventory-rpc/inventory"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetStockLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetStockLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetStockLogic {
	return &GetStockLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetStockLogic) GetStock(in *inventory.GetStockRequest) (*inventory.GetStockResponse, error) {
	inv, err := l.svcCtx.Repository.GetStock(l.ctx, in.ProductId)
	if err != nil {
		return nil, rpcError(err)
	}

	return &inventory.GetStockResponse{
		ProductId: inv.ProductID,
		Stock:     inv.Stock,
	}, nil
}
