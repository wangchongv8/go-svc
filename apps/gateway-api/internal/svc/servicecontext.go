// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package svc

import (
	"go-svc/apps/gateway-api/internal/config"
	userclient "go-svc/apps/user-rpc/userrpc"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config  config.Config
	UserRpc userclient.UserRpc
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:  c,
		UserRpc: userclient.NewUserRpc(zrpc.MustNewClient(c.UserRpcConf)),
	}
}
