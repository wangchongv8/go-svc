// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package auth

import (
	"context"

	"go-svc/apps/gateway-api/internal/svc"
	"go-svc/apps/gateway-api/internal/types"
	userclient "go-svc/apps/user-rpc/userrpc"
	"go-svc/pkg/auth/kratos"
	"go-svc/pkg/observability/traceid"

	"github.com/zeromicro/go-zero/core/logx"
)

type AuthLoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAuthLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AuthLoginLogic {
	return &AuthLoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AuthLoginLogic) AuthLogin(req *types.AuthLoginReq) (resp *types.AuthLoginResp, err error) {
	flowID, err := l.svcCtx.Kratos.InitializeLoginFlow(l.ctx)
	if err != nil {
		return nil, err
	}

	flow, err := l.svcCtx.Kratos.CompleteLoginFlow(l.ctx, flowID, kratos.LoginRequest{
		Identifier: req.Username,
		Password:   req.Password,
		Method:     "password",
	})
	if err != nil {
		return nil, err
	}

	kratosIdentityID := flow.Identity.ID

	userResp, err := l.svcCtx.UserRpc.GetOrCreateByKratosIdentity(traceid.WithOutgoingMetadata(l.ctx), &userclient.GetOrCreateByKratosIdentityRequest{
		KratosIdentityId: kratosIdentityID,
		Username:         req.Username,
	})
	if err != nil {
		l.Errorw("auth login: identity mapping failed",
			logx.Field("kratos_identity_id", kratosIdentityID),
			logx.Field("err", err))
		return nil, err
	}

	l.Infow("auth login success",
		logx.Field("kratos_identity_id", kratosIdentityID),
		logx.Field("user_id", userResp.Id))

	return &types.AuthLoginResp{
		SessionToken: flow.SessionToken,
		UserID:       userResp.Id,
		Username:     userResp.Username,
	}, nil
}
