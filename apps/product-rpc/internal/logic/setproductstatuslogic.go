package logic

import (
	"context"

	"go-svc/apps/product-rpc/internal/svc"
	"go-svc/apps/product-rpc/product"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetProductStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetProductStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetProductStatusLogic {
	return &SetProductStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SetProductStatusLogic) SetProductStatus(in *product.SetProductStatusRequest) (*product.SetProductStatusResponse, error) {
	p, err := l.svcCtx.Repository.SetStatus(l.ctx, in.Id, in.Status)
	if err != nil {
		return nil, rpcError(err)
	}

	return &product.SetProductStatusResponse{
		Id:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		PriceCents:  p.PriceCents,
		Status:      p.Status,
	}, nil
}
