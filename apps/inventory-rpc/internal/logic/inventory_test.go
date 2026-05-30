package logic

import (
	"context"
	"testing"

	"go-svc/apps/inventory-rpc/internal/config"
	"go-svc/apps/inventory-rpc/internal/svc"
	"go-svc/apps/inventory-rpc/inventory"
)

func setupInventoryServiceContext(t *testing.T) *svc.ServiceContext {
	t.Helper()
	return svc.NewServiceContext(config.Config{})
}

func TestSetStock_Success(t *testing.T) {
	svcCtx := setupInventoryServiceContext(t)

	resp, err := NewSetStockLogic(context.Background(), svcCtx).SetStock(&inventory.SetStockRequest{
		ProductId: 1,
		Stock:     100,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Stock != 100 {
		t.Errorf("expected stock 100, got %d", resp.Stock)
	}
}

func TestSetStock_Negative(t *testing.T) {
	svcCtx := setupInventoryServiceContext(t)

	_, err := NewSetStockLogic(context.Background(), svcCtx).SetStock(&inventory.SetStockRequest{
		ProductId: 1,
		Stock:     -1,
	})
	if err == nil {
		t.Fatal("expected error for negative stock")
	}
}

func TestGetStock_NotFound(t *testing.T) {
	svcCtx := setupInventoryServiceContext(t)

	_, err := NewGetStockLogic(context.Background(), svcCtx).GetStock(&inventory.GetStockRequest{ProductId: 999})
	if err == nil {
		t.Fatal("expected error for nonexistent inventory")
	}
}

func TestDeductStock_Success(t *testing.T) {
	svcCtx := setupInventoryServiceContext(t)

	NewSetStockLogic(context.Background(), svcCtx).SetStock(&inventory.SetStockRequest{
		ProductId: 1, Stock: 10,
	})

	resp, err := NewDeductStockLogic(context.Background(), svcCtx).DeductStock(&inventory.DeductStockRequest{
		ProductId: 1,
		Quantity:  3,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Stock != 7 {
		t.Errorf("expected stock 7, got %d", resp.Stock)
	}
}

func TestDeductStock_Insufficient(t *testing.T) {
	svcCtx := setupInventoryServiceContext(t)

	NewSetStockLogic(context.Background(), svcCtx).SetStock(&inventory.SetStockRequest{
		ProductId: 1, Stock: 5,
	})

	_, err := NewDeductStockLogic(context.Background(), svcCtx).DeductStock(&inventory.DeductStockRequest{
		ProductId: 1,
		Quantity:  10,
	})
	if err == nil {
		t.Fatal("expected error for stock insufficient")
	}
}

func TestDeductStock_InvalidQuantity(t *testing.T) {
	svcCtx := setupInventoryServiceContext(t)

	_, err := NewDeductStockLogic(context.Background(), svcCtx).DeductStock(&inventory.DeductStockRequest{
		ProductId: 1,
		Quantity:  0,
	})
	if err == nil {
		t.Fatal("expected error for zero quantity")
	}
}
