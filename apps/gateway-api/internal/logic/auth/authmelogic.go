// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package auth

import (
	"context"
	"fmt"

	"go-svc/apps/gateway-api/internal/authctx"
	"go-svc/apps/gateway-api/internal/svc"
	"go-svc/apps/gateway-api/internal/types"
	userclient "go-svc/apps/user-rpc/userrpc"
	"go-svc/pkg/observability/traceid"

	"github.com/zeromicro/go-zero/core/logx"
)

type AuthMeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAuthMeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AuthMeLogic {
	return &AuthMeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// AuthMe returns the current user profile. It relies on AuthMiddleware
// having already injected user_id into the request context.
func (l *AuthMeLogic) AuthMe() (resp *types.AuthMeResp, err error) {
	userID := authctx.UserIDFromContext(l.ctx)
	if userID == 0 {
		return nil, fmt.Errorf("user not authenticated")
	}

	userResp, err := l.svcCtx.UserRpc.GetUser(traceid.WithOutgoingMetadata(l.ctx), &userclient.GetUserRequest{
		Id: userID,
	})
	if err != nil {
		l.Errorw("auth me: user lookup failed",
			logx.Field("user_id", userID),
			logx.Field("err", err))
		return nil, err
	}

	l.Infow("auth me success",
		logx.Field("user_id", userResp.Id))

	return &types.AuthMeResp{
		UserID:   userResp.Id,
		Username: userResp.Username,
	}, nil
}
