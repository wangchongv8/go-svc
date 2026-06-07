package logic

import (
	"context"

	"go-svc/apps/product-rpc/internal/svc"
	"go-svc/apps/product-rpc/product"
	"go-svc/pkg/observability/traceid"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateProductLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateProductLogic {
	return &CreateProductLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateProductLogic) CreateProduct(in *product.CreateProductRequest) (*product.CreateProductResponse, error) {
	p, err := l.svcCtx.Repository.Create(l.ctx, in.Name, in.Description, in.PriceCents)
	if err != nil {
		return nil, rpcError(err)
	}

	tid, sid := traceid.FromContext(l.ctx)
	fields := []logx.LogField{
		logx.Field("product_id", p.ID),
		logx.Field("name", p.Name),
	}
	if tid != "" {
		fields = append(fields, logx.Field("trace_id", tid), logx.Field("span_id", sid))
	}
	l.Infow("product created", fields...)

	return &product.CreateProductResponse{
		Id:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		PriceCents:  p.PriceCents,
		Status:      p.Status,
	}, nil
}
