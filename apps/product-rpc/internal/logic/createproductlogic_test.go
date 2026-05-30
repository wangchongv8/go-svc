package logic

import (
	"context"
	"testing"

	"go-svc/apps/product-rpc/internal/config"
	"go-svc/apps/product-rpc/internal/svc"
	"go-svc/apps/product-rpc/product"
)

func setupProductServiceContext(t *testing.T) *svc.ServiceContext {
	t.Helper()
	return svc.NewServiceContext(config.Config{})
}

func TestCreateProduct_Success(t *testing.T) {
	svcCtx := setupProductServiceContext(t)

	resp, err := NewCreateProductLogic(context.Background(), svcCtx).CreateProduct(&product.CreateProductRequest{
		Name:        "Keyboard",
		Description: "Mechanical keyboard",
		PriceCents:  19900,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Id == 0 {
		t.Error("expected non-zero id")
	}
	if resp.Name != "Keyboard" {
		t.Errorf("expected name %q, got %q", "Keyboard", resp.Name)
	}
	if resp.Status != "active" {
		t.Errorf("expected status %q, got %q", "active", resp.Status)
	}
}

func TestCreateProduct_InvalidPrice(t *testing.T) {
	svcCtx := setupProductServiceContext(t)

	_, err := NewCreateProductLogic(context.Background(), svcCtx).CreateProduct(&product.CreateProductRequest{
		Name:       "KB",
		PriceCents: 0,
	})
	if err == nil {
		t.Fatal("expected error for zero price")
	}
}

func TestCreateProduct_EmptyName(t *testing.T) {
	svcCtx := setupProductServiceContext(t)

	_, err := NewCreateProductLogic(context.Background(), svcCtx).CreateProduct(&product.CreateProductRequest{
		Name:       "",
		PriceCents: 100,
	})
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestGetProduct_NotFound(t *testing.T) {
	svcCtx := setupProductServiceContext(t)

	_, err := NewGetProductLogic(context.Background(), svcCtx).GetProduct(&product.GetProductRequest{Id: 999})
	if err == nil {
		t.Fatal("expected error for nonexistent product")
	}
}

func TestSetProductStatus_InvalidStatus(t *testing.T) {
	svcCtx := setupProductServiceContext(t)

	resp, _ := NewCreateProductLogic(context.Background(), svcCtx).CreateProduct(&product.CreateProductRequest{
		Name: "KB", PriceCents: 100,
	})

	_, err := NewSetProductStatusLogic(context.Background(), svcCtx).SetProductStatus(&product.SetProductStatusRequest{
		Id:     resp.Id,
		Status: "nonexistent",
	})
	if err == nil {
		t.Fatal("expected error for invalid status")
	}
}

func TestSetProductStatus_Success(t *testing.T) {
	svcCtx := setupProductServiceContext(t)

	resp, _ := NewCreateProductLogic(context.Background(), svcCtx).CreateProduct(&product.CreateProductRequest{
		Name: "KB", PriceCents: 100,
	})

	updated, err := NewSetProductStatusLogic(context.Background(), svcCtx).SetProductStatus(&product.SetProductStatusRequest{
		Id:     resp.Id,
		Status: "inactive",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updated.Status != "inactive" {
		t.Errorf("expected status inactive, got %q", updated.Status)
	}
}

func TestListProducts(t *testing.T) {
	svcCtx := setupProductServiceContext(t)

	NewCreateProductLogic(context.Background(), svcCtx).CreateProduct(&product.CreateProductRequest{Name: "a", PriceCents: 100})
	NewCreateProductLogic(context.Background(), svcCtx).CreateProduct(&product.CreateProductRequest{Name: "b", PriceCents: 200})

	resp, err := NewListProductsLogic(context.Background(), svcCtx).ListProducts(&product.ListProductsRequest{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(resp.Products) != 2 {
		t.Errorf("expected 2 products, got %d", len(resp.Products))
	}
}
