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

type AuthRegisterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAuthRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AuthRegisterLogic {
	return &AuthRegisterLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AuthRegisterLogic) AuthRegister(req *types.AuthRegisterReq) (resp *types.AuthRegisterResp, err error) {
	flowID, err := l.svcCtx.Kratos.InitializeRegistrationFlow(l.ctx)
	if err != nil {
		return nil, err
	}

	flow, err := l.svcCtx.Kratos.CompleteRegistrationFlow(l.ctx, flowID, kratos.RegistrationRequest{
		Traits: struct {
			Username string `json:"username"`
		}{Username: req.Username},
		Password: req.Password,
		Method:   "password",
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
		l.Errorw("auth register: identity mapping failed",
			logx.Field("kratos_identity_id", kratosIdentityID),
			logx.Field("err", err))
		return nil, err
	}

	l.Infow("auth register success",
		logx.Field("kratos_identity_id", kratosIdentityID),
		logx.Field("user_id", userResp.Id))

	return &types.AuthRegisterResp{
		SessionToken: flow.SessionToken,
		UserID:       userResp.Id,
		Username:     userResp.Username,
	}, nil
}
