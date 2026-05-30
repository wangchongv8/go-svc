package product

import (
	"context"

	"go-svc/apps/gateway-api/internal/svc"
	"go-svc/apps/gateway-api/internal/types"
	productclient "go-svc/apps/product-rpc/productrpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListProductsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListProductsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListProductsLogic {
	return &ListProductsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListProductsLogic) ListProducts() (resp *types.ListProductsResp, err error) {
	rpcResp, err := l.svcCtx.ProductRpc.ListProducts(l.ctx, &productclient.ListProductsRequest{})
	if err != nil {
		return nil, err
	}

	result := &types.ListProductsResp{}
	for _, p := range rpcResp.Products {
		result.Products = append(result.Products, types.ProductResp{
			ID:          p.Id,
			Name:        p.Name,
			Description: p.Description,
			PriceCents:  p.PriceCents,
			Status:      p.Status,
		})
	}
	return result, nil
}
