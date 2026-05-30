package model

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	ErrOrderNotFound = errors.New("order not found")
)

type Order struct {
	ID              int64
	UserID          int64
	ProductID       int64
	Quantity        int64
	UnitPriceCents  int64
	TotalPriceCents int64
	Status          string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// OrderRepository defines the data access interface for orders.
type OrderRepository interface {
	Create(ctx context.Context, o *Order) (*Order, error)
	FindByID(ctx context.Context, id int64) (*Order, error)
	ListByUserID(ctx context.Context, userID int64) ([]*Order, error)
}

// ---------------------------------------------------------------
// Fake implementation (in-memory, for testing)
// ---------------------------------------------------------------

type FakeOrderRepository struct {
	mu     sync.RWMutex
	orders map[int64]*Order
	nextID int64
}

func NewFakeOrderRepository() *FakeOrderRepository {
	return &FakeOrderRepository{
		orders: make(map[int64]*Order),
		nextID: 1,
	}
}

func (r *FakeOrderRepository) Create(ctx context.Context, o *Order) (*Order, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	o.ID = r.nextID
	r.nextID++
	o.Status = "created"
	o.CreatedAt = time.Now()
	o.UpdatedAt = time.Now()
	r.orders[o.ID] = o
	return o, nil
}

func (r *FakeOrderRepository) FindByID(ctx context.Context, id int64) (*Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	o, ok := r.orders[id]
	if !ok {
		return nil, ErrOrderNotFound
	}
	return o, nil
}

func (r *FakeOrderRepository) ListByUserID(ctx context.Context, userID int64) ([]*Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*Order
	for _, o := range r.orders {
		if o.UserID == userID {
			result = append(result, o)
		}
	}
	return result, nil
}
