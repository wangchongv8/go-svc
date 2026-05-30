// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package svc

import (
	"go-svc/apps/gateway-api/internal/config"
	inventoryclient "go-svc/apps/inventory-rpc/inventoryrpc"
	orderclient "go-svc/apps/order-rpc/orderrpc"
	productclient "go-svc/apps/product-rpc/productrpc"
	userclient "go-svc/apps/user-rpc/userrpc"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config       config.Config
	UserRpc      userclient.UserRpc
	ProductRpc   productclient.ProductRpc
	InventoryRpc inventoryclient.InventoryRpc
	OrderRpc     orderclient.OrderRpc
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:       c,
		UserRpc:      userclient.NewUserRpc(zrpc.MustNewClient(c.UserRpcConf)),
		ProductRpc:   productclient.NewProductRpc(zrpc.MustNewClient(c.ProductRpcConf)),
		InventoryRpc: inventoryclient.NewInventoryRpc(zrpc.MustNewClient(c.InventoryRpcConf)),
		OrderRpc:     orderclient.NewOrderRpc(zrpc.MustNewClient(c.OrderRpcConf)),
	}
}
