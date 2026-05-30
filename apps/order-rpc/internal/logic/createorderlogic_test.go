package logic

import (
	"context"
	"testing"

	inventoryclient "go-svc/apps/inventory-rpc/inventoryrpc"
	"go-svc/apps/order-rpc/internal/svc"
	"go-svc/apps/order-rpc/model"
	"go-svc/apps/order-rpc/order"
	productclient "go-svc/apps/product-rpc/productrpc"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ---------------------------------------------------------------
// Fake implementations of the RPC client interfaces
// ---------------------------------------------------------------

type fakeProductRpc struct {
	getProductResp *productclient.GetProductResponse
	getProductErr  error
}

func (f *fakeProductRpc) CreateProduct(ctx context.Context, in *productclient.CreateProductRequest, opts ...grpc.CallOption) (*productclient.CreateProductResponse, error) {
	return nil, nil
}
func (f *fakeProductRpc) GetProduct(ctx context.Context, in *productclient.GetProductRequest, opts ...grpc.CallOption) (*productclient.GetProductResponse, error) {
	return f.getProductResp, f.getProductErr
}
func (f *fakeProductRpc) ListProducts(ctx context.Context, in *productclient.ListProductsRequest, opts ...grpc.CallOption) (*productclient.ListProductsResponse, error) {
	return nil, nil
}
func (f *fakeProductRpc) SetProductStatus(ctx context.Context, in *productclient.SetProductStatusRequest, opts ...grpc.CallOption) (*productclient.SetProductStatusResponse, error) {
	return nil, nil
}

var _ productclient.ProductRpc = (*fakeProductRpc)(nil)

type fakeInventoryRpc struct {
	deductStockResp *inventoryclient.DeductStockResponse
	deductStockErr  error
}

func (f *fakeInventoryRpc) SetStock(ctx context.Context, in *inventoryclient.SetStockRequest, opts ...grpc.CallOption) (*inventoryclient.SetStockResponse, error) {
	return nil, nil
}
func (f *fakeInventoryRpc) GetStock(ctx context.Context, in *inventoryclient.GetStockRequest, opts ...grpc.CallOption) (*inventoryclient.GetStockResponse, error) {
	return nil, nil
}
func (f *fakeInventoryRpc) DeductStock(ctx context.Context, in *inventoryclient.DeductStockRequest, opts ...grpc.CallOption) (*inventoryclient.DeductStockResponse, error) {
	return f.deductStockResp, f.deductStockErr
}

var _ inventoryclient.InventoryRpc = (*fakeInventoryRpc)(nil)

// ---------------------------------------------------------------
// Tests
// ---------------------------------------------------------------

func TestCreateOrder_Success(t *testing.T) {
	productRpc := &fakeProductRpc{
		getProductResp: &productclient.GetProductResponse{
			Id: 1, Name: "Keyboard", PriceCents: 19900, Status: "active",
		},
	}
	inventoryRpc := &fakeInventoryRpc{
		deductStockResp: &inventoryclient.DeductStockResponse{ProductId: 1, Stock: 7},
	}
	svcCtx := svc.NewTestServiceContext(
		model.NewFakeOrderRepository(), productRpc, inventoryRpc,
	)

	resp, err := NewCreateOrderLogic(context.Background(), svcCtx).CreateOrder(&order.CreateOrderRequest{
		UserId:    1,
		ProductId: 1,
		Quantity:  2,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Status != "created" {
		t.Errorf("expected status created, got %q", resp.Status)
	}
	if resp.TotalPriceCents != 39800 {
		t.Errorf("expected 39800, got %d", resp.TotalPriceCents)
	}
	if resp.Quantity != 2 {
		t.Errorf("expected quantity 2, got %d", resp.Quantity)
	}

	// Verify order persisted
	saved, err := svcCtx.Repository.FindByID(context.Background(), resp.Id)
	if err != nil {
		t.Fatalf("expected order in repo, got %v", err)
	}
	if saved.TotalPriceCents != 39800 {
		t.Errorf("expected stored 39800, got %d", saved.TotalPriceCents)
	}
}

func TestCreateOrder_ProductNotActive(t *testing.T) {
	productRpc := &fakeProductRpc{
		getProductResp: &productclient.GetProductResponse{
			Id: 1, Name: "Old KB", PriceCents: 100, Status: "inactive",
		},
	}
	inventoryRpc := &fakeInventoryRpc{}
	svcCtx := svc.NewTestServiceContext(
		model.NewFakeOrderRepository(), productRpc, inventoryRpc,
	)

	_, err := NewCreateOrderLogic(context.Background(), svcCtx).CreateOrder(&order.CreateOrderRequest{
		UserId: 1, ProductId: 1, Quantity: 1,
	})
	if err == nil {
		t.Fatal("expected error for inactive product")
	}
	if st, ok := status.FromError(err); !ok || st.Code() != codes.FailedPrecondition {
		t.Errorf("expected FailedPrecondition for inactive product, got %v", err)
	}
}

func TestCreateOrder_StockInsufficient(t *testing.T) {
	productRpc := &fakeProductRpc{
		getProductResp: &productclient.GetProductResponse{
			Id: 1, Name: "KB", PriceCents: 100, Status: "active",
		},
	}
	inventoryRpc := &fakeInventoryRpc{
		deductStockErr: status.Error(codes.FailedPrecondition, "stock insufficient"),
	}
	svcCtx := svc.NewTestServiceContext(
		model.NewFakeOrderRepository(), productRpc, inventoryRpc,
	)

	_, err := NewCreateOrderLogic(context.Background(), svcCtx).CreateOrder(&order.CreateOrderRequest{
		UserId: 1, ProductId: 1, Quantity: 100,
	})
	if err == nil {
		t.Fatal("expected error for stock insufficient")
	}
}
