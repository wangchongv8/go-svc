// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package svc

import (
	"net/http"

	"go-svc/apps/gateway-api/internal/config"
	"go-svc/apps/gateway-api/internal/middleware"
	inventoryclient "go-svc/apps/inventory-rpc/inventoryrpc"
	orderclient "go-svc/apps/order-rpc/orderrpc"
	productclient "go-svc/apps/product-rpc/productrpc"
	userclient "go-svc/apps/user-rpc/userrpc"
	"go-svc/pkg/auth/kratos"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config       config.Config
	Kratos       *kratos.Client
	UserRpc      userclient.UserRpc
	ProductRpc   productclient.ProductRpc
	InventoryRpc inventoryclient.InventoryRpc
	OrderRpc     orderclient.OrderRpc

	// Phase 9: go-zero route middleware for Kratos authentication.
	AuthMiddleware func(http.HandlerFunc) http.HandlerFunc
}

func NewServiceContext(c config.Config) *ServiceContext {
	mw := middleware.NewAuthMiddleware(kratos.NewClient(c.KratosURL), userclient.NewUserRpc(zrpc.MustNewClient(c.UserRpcConf)))

	return &ServiceContext{
		Config:       c,
		Kratos:       mw.KratosClient(),
		UserRpc:      mw.UserRpcClient(),
		ProductRpc:   productclient.NewProductRpc(zrpc.MustNewClient(c.ProductRpcConf)),
		InventoryRpc: inventoryclient.NewInventoryRpc(zrpc.MustNewClient(c.InventoryRpcConf)),
		OrderRpc:     orderclient.NewOrderRpc(zrpc.MustNewClient(c.OrderRpcConf)),

		AuthMiddleware: mw.Handle,
	}
}
