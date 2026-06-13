package logic

import (
	"context"

	"go-svc/apps/user-rpc/internal/svc"
	"go-svc/apps/user-rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetByKratosIdentityLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetByKratosIdentityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetByKratosIdentityLogic {
	return &GetByKratosIdentityLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetByKratosIdentityLogic) GetByKratosIdentity(in *user.GetByKratosIdentityRequest) (*user.GetByKratosIdentityResponse, error) {
	u, err := l.svcCtx.UserStore.FindByKratosIdentity(in.KratosIdentityId)
	if err != nil {
		return nil, rpcError(err)
	}

	return &user.GetByKratosIdentityResponse{
		Id:       u.ID,
		Username: u.Username,
	}, nil
}
