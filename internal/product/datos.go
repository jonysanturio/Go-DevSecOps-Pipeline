package product

import (
	"errors"
	"strings"

	"github.com/jony/inventario/internal/domain"
)

type CreateProductRequest struct {
	Name  string  `json:"name" example:"Laptop Gamer"`
	Price float64 `json:"price" example:"1500.00"`
	Stock int     `json:"stock" example:"10"`
}

func (r CreateProductRequest) Validate() error {
	if strings.TrimSpace(r.Name) == "" {
		return errors.New("El nombre es obligatorio")
	}
	if r.Price < 0 {
		return errors.New("El precio debe ser mayor a cero")
	}
	if r.Stock < 0 {
		return errors.New("Debe tener al menos un stock")
	}
	return nil
}

type ProductResponse struct {
	ID    int     `json:"id" example:"1"`
	Name  string  `json:"name" example:"Coca Cola"`
	Price float64 `json:"price" example:"10.50"`
	Stock int     `json:"stock" example:"100"`
}

func ToResponse(p *domain.Product) ProductResponse {
	return ProductResponse{
		ID:    p.ID,
		Name:  p.Name,
		Price: p.Price,
		Stock: p.Stock,
	}
}

func ToResponseList(products []domain.Product) []ProductResponse {
	list := make([]ProductResponse, len(products))
	for i, p := range products {
		list[i] = ToResponse(&p)
	}
	return list
}