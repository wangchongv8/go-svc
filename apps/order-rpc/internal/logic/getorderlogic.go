package logic

import (
	"context"

	"go-svc/apps/order-rpc/internal/svc"
	"go-svc/apps/order-rpc/order"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOrderLogic {
	return &GetOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetOrderLogic) GetOrder(in *order.GetOrderRequest) (*order.GetOrderResponse, error) {
	o, err := l.svcCtx.Repository.FindByID(l.ctx, in.Id)
	if err != nil {
		return nil, rpcError(err)
	}

	return &order.GetOrderResponse{
		Id:              o.ID,
		UserId:          o.UserID,
		ProductId:       o.ProductID,
		Quantity:        o.Quantity,
		UnitPriceCents:  o.UnitPriceCents,
		TotalPriceCents: o.TotalPriceCents,
		Status:          o.Status,
		CreatedAt:       o.CreatedAt.Unix(),
	}, nil
}
