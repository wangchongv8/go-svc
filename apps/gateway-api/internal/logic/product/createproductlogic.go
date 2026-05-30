package product

import (
	"context"

	"go-svc/apps/gateway-api/internal/svc"
	"go-svc/apps/gateway-api/internal/types"
	productclient "go-svc/apps/product-rpc/productrpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateProductLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateProductLogic {
	return &CreateProductLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateProductLogic) CreateProduct(req *types.CreateProductReq) (resp *types.ProductResp, err error) {
	rpcResp, err := l.svcCtx.ProductRpc.CreateProduct(l.ctx, &productclient.CreateProductRequest{
		Name:        req.Name,
		Description: req.Description,
		PriceCents:  req.PriceCents,
	})
	if err != nil {
		return nil, err
	}

	return &types.ProductResp{
		ID:          rpcResp.Id,
		Name:        rpcResp.Name,
		Description: rpcResp.Description,
		PriceCents:  rpcResp.PriceCents,
		Status:      rpcResp.Status,
	}, nil
}
