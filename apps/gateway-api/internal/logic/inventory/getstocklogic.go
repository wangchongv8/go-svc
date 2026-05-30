package inventory

import (
	"context"

	"go-svc/apps/gateway-api/internal/svc"
	"go-svc/apps/gateway-api/internal/types"
	inventoryclient "go-svc/apps/inventory-rpc/inventoryrpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetStockLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetStockLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetStockLogic {
	return &GetStockLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetStockLogic) GetStock(req *types.GetStockReq) (resp *types.StockResp, err error) {
	rpcResp, err := l.svcCtx.InventoryRpc.GetStock(l.ctx, &inventoryclient.GetStockRequest{
		ProductId: req.ProductID,
	})
	if err != nil {
		return nil, err
	}

	return &types.StockResp{
		ProductID: rpcResp.ProductId,
		Stock:     rpcResp.Stock,
	}, nil
}
