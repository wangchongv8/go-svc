// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf
	KratosURL        string
	UserRpcConf      zrpc.RpcClientConf
	ProductRpcConf   zrpc.RpcClientConf
	InventoryRpcConf zrpc.RpcClientConf
	OrderRpcConf     zrpc.RpcClientConf
}
