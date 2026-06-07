package logic

import (
	"context"

	"go-svc/apps/user-rpc/internal/svc"
	"go-svc/apps/user-rpc/user"
	"go-svc/pkg/observability/traceid"

	"github.com/zeromicro/go-zero/core/logx"
)

type RegisterLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RegisterLogic) Register(in *user.RegisterRequest) (*user.RegisterResponse, error) {
	u, err := l.svcCtx.UserStore.Create(in.Username, in.Password)
	if err != nil {
		return nil, rpcError(err)
	}

	tid, sid := traceid.FromContext(l.ctx)
	fields := []logx.LogField{
		logx.Field("user_id", u.ID),
		logx.Field("username", u.Username),
	}
	if tid != "" {
		fields = append(fields, logx.Field("trace_id", tid), logx.Field("span_id", sid))
	}
	l.Infow("user registered", fields...)

	return &user.RegisterResponse{
		Id:       u.ID,
		Username: u.Username,
	}, nil
}
