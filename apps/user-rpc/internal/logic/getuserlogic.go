package logic

import (
	"context"

	"go-svc/apps/user-rpc/internal/svc"
	"go-svc/apps/user-rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserLogic {
	return &GetUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUserLogic) GetUser(in *user.GetUserRequest) (*user.GetUserResponse, error) {
	u, err := l.svcCtx.UserStore.FindByID(in.Id)
	if err != nil {
		return nil, rpcError(err)
	}

	return &user.GetUserResponse{
		Id:       u.ID,
		Username: u.Username,
	}, nil
}
