package order

import (
	"context"

	"go-svc/apps/gateway-api/internal/svc"
	"go-svc/apps/gateway-api/internal/types"
	orderclient "go-svc/apps/order-rpc/orderrpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOrderLogic {
	return &GetOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetOrderLogic) GetOrder(req *types.GetOrderReq) (resp *types.OrderResp, err error) {
	rpcResp, err := l.svcCtx.OrderRpc.GetOrder(l.ctx, &orderclient.GetOrderRequest{Id: req.ID})
	if err != nil {
		return nil, err
	}

	return &types.OrderResp{
		ID:              rpcResp.Id,
		UserID:          rpcResp.UserId,
		ProductID:       rpcResp.ProductId,
		Quantity:        rpcResp.Quantity,
		UnitPriceCents:  rpcResp.UnitPriceCents,
		TotalPriceCents: rpcResp.TotalPriceCents,
		Status:          rpcResp.Status,
		CreatedAt:       rpcResp.CreatedAt,
	}, nil
}
