package model

import (
	"context"
	"database/sql"
)

// PostgresProductRepository implements ProductRepository backed by PostgreSQL.
type PostgresProductRepository struct {
	db *sql.DB
}

func NewPostgresProductRepository(db *sql.DB) *PostgresProductRepository {
	return &PostgresProductRepository{db: db}
}

func (r *PostgresProductRepository) Create(ctx context.Context, name, description string, priceCents int64) (*Product, error) {
	if name == "" {
		return nil, ErrNameEmpty
	}
	if priceCents <= 0 {
		return nil, ErrInvalidPrice
	}

	row := r.db.QueryRowContext(ctx,
		"INSERT INTO products (name, description, price_cents, status) VALUES ($1, $2, $3, 'active') RETURNING id, name, description, price_cents, status, created_at, updated_at",
		name, description, priceCents)
	return scanProduct(row)
}

func (r *PostgresProductRepository) FindByID(ctx context.Context, id int64) (*Product, error) {
	row := r.db.QueryRowContext(ctx,
		"SELECT id, name, description, price_cents, status, created_at, updated_at FROM products WHERE id = $1", id)
	p, err := scanProduct(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrProductNotFound
		}
		return nil, err
	}
	return p, nil
}

func (r *PostgresProductRepository) List(ctx context.Context) ([]*Product, error) {
	rows, err := r.db.QueryContext(ctx,
		"SELECT id, name, description, price_cents, status, created_at, updated_at FROM products ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []*Product
	for rows.Next() {
		p, err := scanProductRows(rows)
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, rows.Err()
}

func (r *PostgresProductRepository) SetStatus(ctx context.Context, id int64, status string) (*Product, error) {
	if status != "active" && status != "inactive" {
		return nil, ErrInvalidStatus
	}

	row := r.db.QueryRowContext(ctx,
		"UPDATE products SET status = $1, updated_at = now() WHERE id = $2 RETURNING id, name, description, price_cents, status, created_at, updated_at",
		status, id)
	p, err := scanProduct(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrProductNotFound
		}
		return nil, err
	}
	return p, nil
}

func scanProduct(row *sql.Row) (*Product, error) {
	var p Product
	err := row.Scan(&p.ID, &p.Name, &p.Description, &p.PriceCents, &p.Status, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func scanProductRows(rows *sql.Rows) (*Product, error) {
	var p Product
	err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.PriceCents, &p.Status, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}
