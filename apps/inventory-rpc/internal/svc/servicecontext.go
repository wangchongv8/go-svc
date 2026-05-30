package svc

import (
	"context"
	"database/sql"
	"log"
	"time"

	"go-svc/apps/inventory-rpc/internal/config"
	"go-svc/apps/inventory-rpc/model"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type ServiceContext struct {
	Config     config.Config
	Repository model.InventoryRepository
}

func NewServiceContext(c config.Config) *ServiceContext {
	var repo model.InventoryRepository
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
		repo = model.NewPostgresInventoryRepository(db)
	} else {
		repo = model.NewFakeInventoryRepository()
	}

	return &ServiceContext{
		Config:     c,
		Repository: repo,
	}
}
