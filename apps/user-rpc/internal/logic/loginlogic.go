package logic

import (
	"context"

	"go-svc/apps/user-rpc/internal/svc"
	"go-svc/apps/user-rpc/model"
	"go-svc/apps/user-rpc/user"
	"go-svc/pkg/observability/traceid"

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
		tid, _ := traceid.FromContext(l.ctx)
		fields := []logx.LogField{logx.Field("username", in.Username)}
		if tid != "" {
			fields = append(fields, logx.Field("trace_id", tid))
		}
		l.Errorw("login failed: invalid password", fields...)
		return nil, rpcError(model.ErrInvalidPassword)
	}

	tid, sid := traceid.FromContext(l.ctx)
	fields := []logx.LogField{logx.Field("user_id", u.ID)}
	if tid != "" {
		fields = append(fields, logx.Field("trace_id", tid), logx.Field("span_id", sid))
	}
	l.Infow("login success", fields...)
	return &user.LoginResponse{
		Id: u.ID,
	}, nil
}
