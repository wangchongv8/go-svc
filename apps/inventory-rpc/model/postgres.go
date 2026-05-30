package model

import (
	"context"
	"database/sql"
)

// PostgresInventoryRepository implements InventoryRepository backed by PostgreSQL.
type PostgresInventoryRepository struct {
	db *sql.DB
}

func NewPostgresInventoryRepository(db *sql.DB) *PostgresInventoryRepository {
	return &PostgresInventoryRepository{db: db}
}

func (r *PostgresInventoryRepository) SetStock(ctx context.Context, productID, stock int64) (*Inventories, error) {
	if stock < 0 {
		return nil, ErrStockNegative
	}

	row := r.db.QueryRowContext(ctx,
		`INSERT INTO inventories (product_id, stock) VALUES ($1, $2)
		 ON CONFLICT (product_id) DO UPDATE SET stock = $2, updated_at = now()
		 RETURNING product_id, stock`,
		productID, stock)
	var inv Inventories
	if err := row.Scan(&inv.ProductID, &inv.Stock); err != nil {
		return nil, err
	}
	return &inv, nil
}

func (r *PostgresInventoryRepository) GetStock(ctx context.Context, productID int64) (*Inventories, error) {
	row := r.db.QueryRowContext(ctx,
		"SELECT product_id, stock FROM inventories WHERE product_id = $1", productID)
	var inv Inventories
	if err := row.Scan(&inv.ProductID, &inv.Stock); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrInventoryNotFound
		}
		return nil, err
	}
	return &inv, nil
}

func (r *PostgresInventoryRepository) DeductStock(ctx context.Context, productID, quantity int64) (*Inventories, error) {
	if quantity <= 0 {
		return nil, ErrQuantityInvalid
	}

	row := r.db.QueryRowContext(ctx,
		`UPDATE inventories
		 SET stock = stock - $2, updated_at = now()
		 WHERE product_id = $1 AND stock >= $2
		 RETURNING product_id, stock`,
		productID, quantity)
	var inv Inventories
	if err := row.Scan(&inv.ProductID, &inv.Stock); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrStockInsufficient
		}
		return nil, err
	}
	return &inv, nil
}
