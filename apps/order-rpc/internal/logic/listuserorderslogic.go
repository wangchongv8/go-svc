package logic

import (
	"context"

	"go-svc/apps/order-rpc/internal/svc"
	"go-svc/apps/order-rpc/order"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListUserOrdersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListUserOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListUserOrdersLogic {
	return &ListUserOrdersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListUserOrdersLogic) ListUserOrders(in *order.ListUserOrdersRequest) (*order.ListUserOrdersResponse, error) {
	orders, err := l.svcCtx.Repository.ListByUserID(l.ctx, in.UserId)
	if err != nil {
		return nil, rpcError(err)
	}

	resp := &order.ListUserOrdersResponse{}
	for _, o := range orders {
		resp.Orders = append(resp.Orders, &order.GetOrderResponse{
			Id:              o.ID,
			UserId:          o.UserID,
			ProductId:       o.ProductID,
			Quantity:        o.Quantity,
			UnitPriceCents:  o.UnitPriceCents,
			TotalPriceCents: o.TotalPriceCents,
			Status:          o.Status,
			CreatedAt:       o.CreatedAt.Unix(),
		})
	}
	return resp, nil
}
