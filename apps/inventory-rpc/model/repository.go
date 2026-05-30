package model

import (
	"context"
	"errors"
	"sync"
)

var (
	ErrInventoryNotFound = errors.New("inventory not found")
	ErrStockInsufficient = errors.New("stock insufficient")
	ErrQuantityInvalid   = errors.New("quantity must be greater than 0")
	ErrStockNegative     = errors.New("stock must not be negative")
)

type Inventories struct {
	ProductID int64
	Stock     int64
}

// InventoryRepository defines the data access interface for inventory.
type InventoryRepository interface {
	SetStock(ctx context.Context, productID, stock int64) (*Inventories, error)
	GetStock(ctx context.Context, productID int64) (*Inventories, error)
	DeductStock(ctx context.Context, productID, quantity int64) (*Inventories, error)
}

// ---------------------------------------------------------------
// Fake implementation (in-memory, for testing)
// ---------------------------------------------------------------

type FakeInventoryRepository struct {
	mu          sync.RWMutex
	inventories map[int64]*Inventories
}

func NewFakeInventoryRepository() *FakeInventoryRepository {
	return &FakeInventoryRepository{
		inventories: make(map[int64]*Inventories),
	}
}

func (r *FakeInventoryRepository) SetStock(ctx context.Context, productID, stock int64) (*Inventories, error) {
	if stock < 0 {
		return nil, ErrStockNegative
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	inv := &Inventories{ProductID: productID, Stock: stock}
	r.inventories[productID] = inv
	return inv, nil
}

func (r *FakeInventoryRepository) GetStock(ctx context.Context, productID int64) (*Inventories, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	inv, ok := r.inventories[productID]
	if !ok {
		return nil, ErrInventoryNotFound
	}
	return inv, nil
}

func (r *FakeInventoryRepository) DeductStock(ctx context.Context, productID, quantity int64) (*Inventories, error) {
	if quantity <= 0 {
		return nil, ErrQuantityInvalid
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	inv, ok := r.inventories[productID]
	if !ok {
		return nil, ErrInventoryNotFound
	}
	if inv.Stock < quantity {
		return nil, ErrStockInsufficient
	}
	inv.Stock -= quantity
	return inv, nil
}
