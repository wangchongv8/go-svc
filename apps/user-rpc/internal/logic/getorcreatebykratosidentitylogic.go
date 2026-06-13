package logic

import (
	"context"

	"go-svc/apps/user-rpc/internal/svc"
	"go-svc/apps/user-rpc/user"
	"go-svc/pkg/observability/traceid"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetOrCreateByKratosIdentityLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetOrCreateByKratosIdentityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOrCreateByKratosIdentityLogic {
	return &GetOrCreateByKratosIdentityLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetOrCreateByKratosIdentityLogic) GetOrCreateByKratosIdentity(in *user.GetOrCreateByKratosIdentityRequest) (*user.GetOrCreateByKratosIdentityResponse, error) {
	u, err := l.svcCtx.UserStore.GetOrCreateByKratosIdentity(in.KratosIdentityId, in.Username)
	if err != nil {
		return nil, rpcError(err)
	}

	tid, sid := traceid.FromContext(l.ctx)
	fields := []logx.LogField{
		logx.Field("user_id", u.ID),
		logx.Field("username", u.Username),
		logx.Field("kratos_identity_id", in.KratosIdentityId),
	}
	if tid != "" {
		fields = append(fields, logx.Field("trace_id", tid), logx.Field("span_id", sid))
	}
	l.Infow("get or create by kratos identity", fields...)

	return &user.GetOrCreateByKratosIdentityResponse{
		Id:       u.ID,
		Username: u.Username,
	}, nil
}
