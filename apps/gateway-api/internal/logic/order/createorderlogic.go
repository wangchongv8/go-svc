// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package order

import (
	"context"

	"go-svc/apps/gateway-api/internal/svc"
	"go-svc/apps/gateway-api/internal/types"
	orderclient "go-svc/apps/order-rpc/orderrpc"
	"go-svc/pkg/observability/traceid"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateOrderLogic {
	return &CreateOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateOrderLogic) CreateOrder(req *types.CreateOrderReq) (resp *types.OrderResp, err error) {
	rpcResp, err := l.svcCtx.OrderRpc.CreateOrder(traceid.WithOutgoingMetadata(l.ctx), &orderclient.CreateOrderRequest{
		UserId:    req.UserID,
		ProductId: req.ProductID,
		Quantity:  req.Quantity,
	})
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
