package config

import (
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	Dsn              string
	ProductRpcConf   zrpc.RpcClientConf
	InventoryRpcConf zrpc.RpcClientConf
}
