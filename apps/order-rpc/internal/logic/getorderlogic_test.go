package logic

import (
	"context"
	"testing"

	"go-svc/apps/order-rpc/internal/svc"
	"go-svc/apps/order-rpc/model"
	"go-svc/apps/order-rpc/order"
)

// testServiceContext returns a ServiceContext with fake repo and nil RPC clients
// (safe for logic that does not call RPCs, like GetOrder/ListUserOrders).
func testServiceContext() *svc.ServiceContext {
	return svc.NewTestServiceContext(model.NewFakeOrderRepository(), nil, nil)
}

func TestGetOrder_NotFound(t *testing.T) {
	svcCtx := testServiceContext()

	_, err := NewGetOrderLogic(context.Background(), svcCtx).GetOrder(&order.GetOrderRequest{Id: 999})
	if err == nil {
		t.Fatal("expected error for nonexistent order")
	}
}

func TestGetOrder_Success(t *testing.T) {
	svcCtx := testServiceContext()

	o, _ := svcCtx.Repository.Create(context.Background(), &model.Order{
		UserID:          1,
		ProductID:       1,
		Quantity:        2,
		UnitPriceCents:  100,
		TotalPriceCents: 200,
	})

	resp, err := NewGetOrderLogic(context.Background(), svcCtx).GetOrder(&order.GetOrderRequest{Id: o.ID})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Quantity != 2 {
		t.Errorf("expected quantity 2, got %d", resp.Quantity)
	}
}
