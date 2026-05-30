package product

import (
	"context"

	"go-svc/apps/gateway-api/internal/svc"
	"go-svc/apps/gateway-api/internal/types"
	productclient "go-svc/apps/product-rpc/productrpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetProductLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetProductLogic {
	return &GetProductLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetProductLogic) GetProduct(req *types.GetProductReq) (resp *types.ProductResp, err error) {
	rpcResp, err := l.svcCtx.ProductRpc.GetProduct(l.ctx, &productclient.GetProductRequest{Id: req.ID})
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
