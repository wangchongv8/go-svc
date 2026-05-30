package inventory

import (
	"context"

	"go-svc/apps/gateway-api/internal/svc"
	"go-svc/apps/gateway-api/internal/types"
	inventoryclient "go-svc/apps/inventory-rpc/inventoryrpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetStockLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSetStockLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetStockLogic {
	return &SetStockLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SetStockLogic) SetStock(req *types.SetStockReq) (resp *types.StockResp, err error) {
	rpcResp, err := l.svcCtx.InventoryRpc.SetStock(l.ctx, &inventoryclient.SetStockRequest{
		ProductId: req.ProductID,
		Stock:     req.Stock,
	})
	if err != nil {
		return nil, err
	}

	return &types.StockResp{
		ProductID: rpcResp.ProductId,
		Stock:     rpcResp.Stock,
	}, nil
}
