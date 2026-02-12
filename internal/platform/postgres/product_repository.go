package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jony/inventario/internal/domain"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// 1. Guardar (Save)
func (r *Repository) Save(ctx context.Context, p *domain.Product) error {
	query := "INSERT INTO products (name, price, stock) VALUES ($1, $2, $3) RETURNING id"
	err := r.db.QueryRowContext(ctx, query, p.Name, p.Price, p.Stock).Scan(&p.ID)
	if err != nil {
		return fmt.Errorf("error saving product: %w", err)
	}
	return nil
}

// 2. Traer Todos (GetAll)
func (r *Repository) GetAll(ctx context.Context) ([]domain.Product, error) {
	query := "SELECT id, name, price, stock FROM products"
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error fetching products: %w", err)
	}
	defer rows.Close()

	var products []domain.Product
	for rows.Next() {
		var p domain.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Price, &p.Stock); err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, nil
} 
// <--- ¡IMPORTANTE! Esta llave cierra GetAll. Si te faltaba, GetOne quedaba adentro.

// 3. Traer Uno (GetOne) - ¡Aquí estaba el problema!
func (r *Repository) GetOne(ctx context.Context, id int) (*domain.Product, error) {
	query := "SELECT id, name, price, stock FROM products WHERE id = $1"
	
	row := r.db.QueryRowContext(ctx, query, id)

	var p domain.Product
	err := row.Scan(&p.ID, &p.Name, &p.Price, &p.Stock)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// FUNCION DE UPDATE PARA LOS PRODUCTOS QUE YA EXISTEN 
func (r *Repository) Update(ctx context.Context, id int, p *domain.Product)error{
	query := "UPDATE products SET name = $1, price = $2, stock = $3 WHERE id = $4"
	// EJECUCION DE 3 DATOS + EL ID PARA FILTRARLO
	_, err := r.db.ExecContext(ctx, query, p.Name, p.Stock, p.Price, id)
	if err != nil{
		return fmt.Errorf("ERROR UPDATING PRODUCT: %w", err)
	}
	return nil
}
func (r *Repository)Delete(ctx context.Context, id int)error{
	query := "DELETE FROM products where id = $1"
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil{
		return fmt.Errorf("ERROR DELATING PRODUCT: %w", err)
	}
	return nil
}
