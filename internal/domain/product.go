package domain

import (
	"errors"
	"context"
)

var (
	ErrProductNotFound = errors.New("Product Not Found")
	ErrInvalidPrice = errors.New("Price cannot be negative")
)


type Product struct {
	ID		int 	`json:"id"`
	Name	string	`json:"name"`
	Price	float64	`json:"price"`
	Stock	int		`json:"stock"` 
}


func (p *Product) Validate() error{
	if p.Name == ""{
	return errors.New("Product name is required")
	}
	if p.Price < 0 {
		return ErrInvalidPrice
	}
	return nil
}



type ProductRepository interface {
	Save(ctx context.Context, p *Product) error
	GetAll(ctx context.Context) ([]Product, error)
	GetOne(ctx context.Context, d int) (*Product, error)
	Update(ctx context.Context, id int, p *Product) error
	Delete(ctx context.Context, id int) error
}