package logic

import (
	"context"

	"go-svc/apps/user-rpc/internal/svc"
	"go-svc/apps/user-rpc/model"
	"go-svc/apps/user-rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *LoginLogic) Login(in *user.LoginRequest) (*user.LoginResponse, error) {
	u, err := l.svcCtx.UserStore.FindByUsername(in.Username)
	if err != nil {
		return nil, rpcError(err)
	}

	if u.Password != in.Password {
		l.Errorw("login failed: invalid password", logx.Field("username", in.Username))
		return nil, rpcError(model.ErrInvalidPassword)
	}

	l.Infow("login success", logx.Field("user_id", u.ID))
	return &user.LoginResponse{
		Id: u.ID,
	}, nil
}
