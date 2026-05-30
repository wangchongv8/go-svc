package logic

import (
	"context"

	"go-svc/apps/product-rpc/internal/svc"
	"go-svc/apps/product-rpc/product"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListProductsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListProductsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListProductsLogic {
	return &ListProductsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListProductsLogic) ListProducts(in *product.ListProductsRequest) (*product.ListProductsResponse, error) {
	products, err := l.svcCtx.Repository.List(l.ctx)
	if err != nil {
		return nil, rpcError(err)
	}

	resp := &product.ListProductsResponse{}
	for _, p := range products {
		resp.Products = append(resp.Products, &product.GetProductResponse{
			Id:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			PriceCents:  p.PriceCents,
			Status:      p.Status,
		})
	}
	return resp, nil
}
