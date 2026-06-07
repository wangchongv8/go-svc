package logic

import (
	"context"

	inventoryclient "go-svc/apps/inventory-rpc/inventoryrpc"
	"go-svc/apps/order-rpc/internal/svc"
	"go-svc/apps/order-rpc/model"
	"go-svc/apps/order-rpc/order"
	"go-svc/pkg/observability/traceid"
	productclient "go-svc/apps/product-rpc/productrpc"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CreateOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateOrderLogic {
	return &CreateOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateOrderLogic) CreateOrder(in *order.CreateOrderRequest) (*order.CreateOrderResponse, error) {
	if in.UserId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id must be greater than 0")
	}

	product, err := l.svcCtx.ProductRpc.GetProduct(l.ctx, &productclient.GetProductRequest{Id: in.ProductId})
	if err != nil {
		return nil, err
	}
	if product.Status != "active" {
		return nil, status.Error(codes.FailedPrecondition, "product is not active")
	}

	_, err = l.svcCtx.InventoryRpc.DeductStock(l.ctx, &inventoryclient.DeductStockRequest{
		ProductId: in.ProductId,
		Quantity:  in.Quantity,
	})
	if err != nil {
		etid, esid := traceid.FromContext(l.ctx)
		l.Errorw("stock deduct failed",
			logx.Field("product_id", in.ProductId),
			logx.Field("user_id", in.UserId),
			logx.Field("trace_id", etid),
			logx.Field("span_id", esid))
		return nil, err
	}

	o := &model.Order{
		UserID:          in.UserId,
		ProductID:       in.ProductId,
		Quantity:        in.Quantity,
		UnitPriceCents:  product.PriceCents,
		TotalPriceCents: product.PriceCents * in.Quantity,
	}
	created, err := l.svcCtx.Repository.Create(l.ctx, o)
	if err != nil {
		return nil, rpcError(err)
	}

	tid, sid := traceid.FromContext(l.ctx)

	l.Infow("order created",
		logx.Field("order_id", created.ID),
		logx.Field("user_id", created.UserID),
		logx.Field("product_id", created.ProductID),
		logx.Field("total_price_cents", created.TotalPriceCents),
		logx.Field("trace_id", tid),
		logx.Field("span_id", sid))

	return toOrderResponse(created), nil
}

func toOrderResponse(o *model.Order) *order.CreateOrderResponse {
	return &order.CreateOrderResponse{
		Id:              o.ID,
		UserId:          o.UserID,
		ProductId:       o.ProductID,
		Quantity:        o.Quantity,
		UnitPriceCents:  o.UnitPriceCents,
		TotalPriceCents: o.TotalPriceCents,
		Status:          o.Status,
		CreatedAt:       o.CreatedAt.Unix(),
	}
}
