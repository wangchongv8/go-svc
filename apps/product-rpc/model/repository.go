package model

import (
	"context"
	"errors"
	"sort"
	"sync"
	"time"
)

var (
	ErrProductNotFound = errors.New("product not found")
	ErrInvalidPrice    = errors.New("price_cents must be greater than 0")
	ErrInvalidStatus   = errors.New("status must be active or inactive")
	ErrNameEmpty       = errors.New("product name must not be empty")
)

type Product struct {
	ID          int64
	Name        string
	Description string
	PriceCents  int64
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ProductRepository defines the data access interface for products.
type ProductRepository interface {
	Create(ctx context.Context, name, description string, priceCents int64) (*Product, error)
	FindByID(ctx context.Context, id int64) (*Product, error)
	List(ctx context.Context) ([]*Product, error)
	SetStatus(ctx context.Context, id int64, status string) (*Product, error)
}

// ---------------------------------------------------------------
// Fake implementation (in-memory, for testing)
// ---------------------------------------------------------------

type FakeProductRepository struct {
	mu       sync.RWMutex
	products map[int64]*Product
	nextID   int64
}

func NewFakeProductRepository() *FakeProductRepository {
	return &FakeProductRepository{
		products: make(map[int64]*Product),
		nextID:   1,
	}
}

func (r *FakeProductRepository) Create(ctx context.Context, name, description string, priceCents int64) (*Product, error) {
	if name == "" {
		return nil, ErrNameEmpty
	}
	if priceCents <= 0 {
		return nil, ErrInvalidPrice
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	p := &Product{
		ID:          r.nextID,
		Name:        name,
		Description: description,
		PriceCents:  priceCents,
		Status:      "active",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	r.nextID++
	r.products[p.ID] = p
	return p, nil
}

func (r *FakeProductRepository) FindByID(ctx context.Context, id int64) (*Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, ok := r.products[id]
	if !ok {
		return nil, ErrProductNotFound
	}
	return p, nil
}

func (r *FakeProductRepository) List(ctx context.Context) ([]*Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	products := make([]*Product, 0, len(r.products))
	for _, p := range r.products {
		products = append(products, p)
	}
	sort.Slice(products, func(i, j int) bool { return products[i].ID < products[j].ID })
	return products, nil
}

func (r *FakeProductRepository) SetStatus(ctx context.Context, id int64, status string) (*Product, error) {
	if status != "active" && status != "inactive" {
		return nil, ErrInvalidStatus
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	p, ok := r.products[id]
	if !ok {
		return nil, ErrProductNotFound
	}
	p.Status = status
	p.UpdatedAt = time.Now()
	return p, nil
}
