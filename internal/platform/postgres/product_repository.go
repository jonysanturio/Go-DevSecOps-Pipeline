package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jony/inventario/internal/domain"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Save(ctx context.Context, p *domain.Product) error {
	query := "INSERT INTO products (name, price, stock) VALUES ($1, $2, $3) RETURNING id"
	err := r.db.QueryRowContext(ctx, query, p.Name, p.Price, p.Stock).Scan(&p.ID)
	if err != nil {
		return fmt.Errorf("error saving product: %w", err)
	}
	return nil
}

func (r *Repository) GetAll(ctx context.Context) (products []domain.Product, retErr error) {
	query := "SELECT id, name, price, stock FROM products"
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error fetching products: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			retErr = errors.Join(retErr, fmt.Errorf("close product rows: %w", closeErr))
		}
	}()

	for rows.Next() {
		var p domain.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Price, &p.Stock); err != nil {
			return nil, fmt.Errorf("error scanning product: %w", err)
		}
		products = append(products, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating products: %w", err)
	}
	return products, nil
}

func (r *Repository) GetOne(ctx context.Context, id int) (*domain.Product, error) {
	query := "SELECT id, name, price, stock FROM products WHERE id = $1"

	row := r.db.QueryRowContext(ctx, query, id)

	var p domain.Product
	err := row.Scan(&p.ID, &p.Name, &p.Price, &p.Stock)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrProductNotFound
		}
		return nil, fmt.Errorf("error fetching product: %w", err)
	}
	return &p, nil
}

func (r *Repository) Update(ctx context.Context, id int, p *domain.Product) error {
	query := `UPDATE products
		SET name = $1, price = $2, stock = $3
		WHERE id = $4`
	result, err := r.db.ExecContext(ctx, query, p.Name, p.Price, p.Stock, id)
	if err != nil {
		return fmt.Errorf("error updating product: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking updated product: %w", err)
	}
	if rowsAffected == 0 {
		return domain.ErrProductNotFound
	}
	return nil
}
func (r *Repository) Delete(ctx context.Context, id int) error {
	query := "DELETE FROM products WHERE id = $1"
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error deleting product: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking deleted product: %w", err)
	}
	if rowsAffected == 0 {
		return domain.ErrProductNotFound
	}
	return nil
}
