// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package user

import (
	"context"

	"go-svc/apps/gateway-api/internal/svc"
	"go-svc/apps/gateway-api/internal/types"
	userclient "go-svc/apps/user-rpc/userrpc"
	"go-svc/pkg/observability/traceid"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserLogic {
	return &GetUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserLogic) GetUser(req *types.GetUserReq) (resp *types.GetUserResp, err error) {
	rpcResp, err := l.svcCtx.UserRpc.GetUser(traceid.WithOutgoingMetadata(l.ctx), &userclient.GetUserRequest{
		Id: req.ID,
	})
	if err != nil {
		return nil, err
	}

	return &types.GetUserResp{
		ID:       rpcResp.Id,
		Username: rpcResp.Username,
	}, nil
}
