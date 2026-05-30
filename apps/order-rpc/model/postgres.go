package model

import (
	"context"
	"database/sql"
)

// PostgresOrderRepository implements OrderRepository backed by PostgreSQL.
type PostgresOrderRepository struct {
	db *sql.DB
}

func NewPostgresOrderRepository(db *sql.DB) *PostgresOrderRepository {
	return &PostgresOrderRepository{db: db}
}

func (r *PostgresOrderRepository) Create(ctx context.Context, o *Order) (*Order, error) {
	row := r.db.QueryRowContext(ctx,
		`INSERT INTO orders (user_id, product_id, quantity, unit_price_cents, total_price_cents, status)
		 VALUES ($1, $2, $3, $4, $5, 'created')
		 RETURNING id, user_id, product_id, quantity, unit_price_cents, total_price_cents, status, created_at, updated_at`,
		o.UserID, o.ProductID, o.Quantity, o.UnitPriceCents, o.TotalPriceCents)
	var result Order
	err := row.Scan(&result.ID, &result.UserID, &result.ProductID, &result.Quantity,
		&result.UnitPriceCents, &result.TotalPriceCents, &result.Status, &result.CreatedAt, &result.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *PostgresOrderRepository) FindByID(ctx context.Context, id int64) (*Order, error) {
	row := r.db.QueryRowContext(ctx,
		"SELECT id, user_id, product_id, quantity, unit_price_cents, total_price_cents, status, created_at, updated_at FROM orders WHERE id = $1", id)
	var o Order
	err := row.Scan(&o.ID, &o.UserID, &o.ProductID, &o.Quantity,
		&o.UnitPriceCents, &o.TotalPriceCents, &o.Status, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}
	return &o, nil
}

func (r *PostgresOrderRepository) ListByUserID(ctx context.Context, userID int64) ([]*Order, error) {
	rows, err := r.db.QueryContext(ctx,
		"SELECT id, user_id, product_id, quantity, unit_price_cents, total_price_cents, status, created_at, updated_at FROM orders WHERE user_id = $1 ORDER BY id DESC", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*Order
	for rows.Next() {
		var o Order
		if err := rows.Scan(&o.ID, &o.UserID, &o.ProductID, &o.Quantity,
			&o.UnitPriceCents, &o.TotalPriceCents, &o.Status, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, err
		}
		orders = append(orders, &o)
	}
	return orders, rows.Err()
}
