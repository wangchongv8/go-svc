// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package order

import (
	"context"
	"fmt"

	"go-svc/apps/gateway-api/internal/authctx"
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

// CreateOrder reads user_id from the auth context and forwards to CreateOrderWithUserID.
func (l *CreateOrderLogic) CreateOrder(req *types.CreateOrderReq) (resp *types.OrderResp, err error) {
	userID := authctx.UserIDFromContext(l.ctx)
	if userID == 0 {
		return nil, fmt.Errorf("user not authenticated")
	}
	return l.CreateOrderWithUserID(req.ProductID, req.Quantity, userID)
}

// CreateOrderWithUserID creates an order with an explicit user_id from auth context.
func (l *CreateOrderLogic) CreateOrderWithUserID(productID, quantity, userID int64) (resp *types.OrderResp, err error) {
	rpcResp, err := l.svcCtx.OrderRpc.CreateOrder(traceid.WithOutgoingMetadata(l.ctx), &orderclient.CreateOrderRequest{
		UserId:    userID,
		ProductId: productID,
		Quantity:  quantity,
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
