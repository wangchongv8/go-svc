package svc

import (
	"go-svc/apps/user-rpc/internal/config"
	"go-svc/apps/user-rpc/model"
)

type ServiceContext struct {
	Config    config.Config
	UserStore *model.UserStore
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:    c,
		UserStore: model.NewUserStore(),
	}
}
