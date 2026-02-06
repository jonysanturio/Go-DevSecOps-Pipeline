package product

import (
	"github.com/jony/inventario/internal/domain"
)

// Service: Define QUÉ puedes hacer con los productos.

type Service struct {
	repo domain.ProductRepository
}

// NewService: Constructor. Le inyectamos el repositorio que queramos usar.
func NewService(repo domain.ProductRepository) *Service {
	return &Service{repo: repo}
}

// Create: La lógica completa de crear un producto.

func (s *Service)Create(name string, price float64, stock int)(*domain.Product, error){
	// 1. Instanciamos el producto (El Plano)
	p := &domain.Product{
		Name: name,
		Price: price,
		Stock: stock,
	}
	// 2. Ejecutamos validaciones de negocio (Reglas universales)
	if err := p.Validate(); err != nil{
		return nil, err
	}
	if err := s.repo.Save(p); err != nil{
		return nil, err
	}
	return p, nil
}

// GetAll: Pide todos los productos
func (s *Service) GetAll() ([]domain.Product, error){
	return s.repo.GetAll()
}

func (s *Service) GetOne(id int)(*domain.Product, error){
	return  s.repo.GetOne(id)
}

func (s *Service)Update(id int, name string, price float64, stock int)(*domain.Product, error){
	p := &domain.Product{
		ID: 	id,
		Name:	name,
		Price:	price,
		Stock: 	stock,
	}
	// VALIDAR STOCK 
	if err := p.Validate(); err != nil{
		return nil, err
	}
	// LLAMAR AL REPO
	if err := s.repo.Update(id, p); err != nil{
		return nil, err
	}
	return p, nil
}

func (s *Service)Delete(id int)error{
	return s.repo.Delete(id)
}

// CreateProductRequest es lo que manda el usuario (Input)
type CreateProductRequest struct {
    Name  string  `json:"name" example:"Laptop Gamer"`
    Price float64 `json:"price" example:"1500.00"`
    Stock int     `json:"stock" example:"10"`
}

// Product es lo que devolvemos (Output)
type Product struct {
    ID    int     `json:"id" example:"1"`
    Name  string  `json:"name" example:"Coca Cola"`
    Price float64 `json:"price" example:"10.50"`
    Stock int     `json:"stock" example:"100"`
}