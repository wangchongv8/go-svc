package svc

import (
	"context"
	"database/sql"
	"log"
	"time"

	inventoryclient "go-svc/apps/inventory-rpc/inventoryrpc"
	"go-svc/apps/order-rpc/internal/config"
	"go-svc/apps/order-rpc/model"
	productclient "go-svc/apps/product-rpc/productrpc"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config       config.Config
	Repository   model.OrderRepository
	ProductRpc   productclient.ProductRpc
	InventoryRpc inventoryclient.InventoryRpc
}

func NewServiceContext(c config.Config) *ServiceContext {
	var repo model.OrderRepository
	if c.Dsn != "" {
		db, err := sql.Open("pgx", c.Dsn)
		if err != nil {
			log.Fatalf("failed to open database: %v", err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := db.PingContext(ctx); err != nil {
			log.Fatalf("failed to ping database: %v", err)
		}
		repo = model.NewPostgresOrderRepository(db)
	} else {
		repo = model.NewFakeOrderRepository()
	}

	var productRpc productclient.ProductRpc
	if c.ProductRpcConf.Target != "" {
		productRpc = productclient.NewProductRpc(zrpc.MustNewClient(c.ProductRpcConf))
	}
	var inventoryRpc inventoryclient.InventoryRpc
	if c.InventoryRpcConf.Target != "" {
		inventoryRpc = inventoryclient.NewInventoryRpc(zrpc.MustNewClient(c.InventoryRpcConf))
	}

	return &ServiceContext{
		Config:       c,
		Repository:   repo,
		ProductRpc:   productRpc,
		InventoryRpc: inventoryRpc,
	}
}

// NewTestServiceContext creates a ServiceContext with injected dependencies for testing.
func NewTestServiceContext(
	repo model.OrderRepository,
	productRpc productclient.ProductRpc,
	inventoryRpc inventoryclient.InventoryRpc,
) *ServiceContext {
	return &ServiceContext{
		Config:       config.Config{},
		Repository:   repo,
		ProductRpc:   productRpc,
		InventoryRpc: inventoryRpc,
	}
}
