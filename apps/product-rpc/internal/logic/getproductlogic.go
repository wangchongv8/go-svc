package logic

import (
	"context"

	"go-svc/apps/product-rpc/internal/svc"
	"go-svc/apps/product-rpc/product"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetProductLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetProductLogic {
	return &GetProductLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetProductLogic) GetProduct(in *product.GetProductRequest) (*product.GetProductResponse, error) {
	p, err := l.svcCtx.Repository.FindByID(l.ctx, in.Id)
	if err != nil {
		return nil, rpcError(err)
	}

	return &product.GetProductResponse{
		Id:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		PriceCents:  p.PriceCents,
		Status:      p.Status,
	}, nil
}
