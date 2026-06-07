package product

import (
	"context"

	"go-svc/apps/gateway-api/internal/svc"
	"go-svc/apps/gateway-api/internal/types"
	productclient "go-svc/apps/product-rpc/productrpc"
	"go-svc/pkg/observability/traceid"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetProductStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSetProductStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetProductStatusLogic {
	return &SetProductStatusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SetProductStatusLogic) SetProductStatus(req *types.SetProductStatusReq) (resp *types.ProductResp, err error) {
	rpcResp, err := l.svcCtx.ProductRpc.SetProductStatus(traceid.WithOutgoingMetadata(l.ctx), &productclient.SetProductStatusRequest{
		Id:     req.ID,
		Status: req.Status,
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
