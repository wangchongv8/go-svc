package order

import (
	"context"

	"go-svc/apps/gateway-api/internal/svc"
	"go-svc/apps/gateway-api/internal/types"
	orderclient "go-svc/apps/order-rpc/orderrpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListUserOrdersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListUserOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListUserOrdersLogic {
	return &ListUserOrdersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListUserOrdersLogic) ListUserOrders(req *types.ListUserOrdersReq) (resp *types.ListUserOrdersResp, err error) {
	rpcResp, err := l.svcCtx.OrderRpc.ListUserOrders(l.ctx, &orderclient.ListUserOrdersRequest{
		UserId: req.UserID,
	})
	if err != nil {
		return nil, err
	}

	result := &types.ListUserOrdersResp{}
	for _, o := range rpcResp.Orders {
		result.Orders = append(result.Orders, types.OrderResp{
			ID:              o.Id,
			UserID:          o.UserId,
			ProductID:       o.ProductId,
			Quantity:        o.Quantity,
			UnitPriceCents:  o.UnitPriceCents,
			TotalPriceCents: o.TotalPriceCents,
			Status:          o.Status,
			CreatedAt:       o.CreatedAt,
		})
	}
	return result, nil
}
