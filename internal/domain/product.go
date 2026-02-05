package domain

import "errors"

//Errores de los dominios (Regla de negocio)

var (
	ErrProductNotFound = errors.New("Product Not Found")
	ErrInvalidPrice = errors.New("Price cannot be negative")
)

// Prodcut: Es la entidad pura

type Product struct {
	ID		int 	`json:"id"`
	Name	string	`json:"name"`
	Price	float64	`json:"price"`
	Stock	int		`json:"stock"` 
}

// Validate: Metodo para asegurar que el dato es valido antes de procesarlo
// Lo define el Defense Programming
func (p *Product) Validate() error{
	if p.Name == ""{
	return errors.New("Product name is required")
	}
	if p.Price < 0 {
		return ErrInvalidPrice
	}
	return nil
}

// ProductRepository
// Es una interfaz (osea un contrato)
// Le estamos diciendo al mundo: "Necesito a alguien que sepa Guardar y Buscar productos".
// No me importa si es Postgres o Mongo. Solo quiero que cumpla este contrato.

type ProductRepository interface {
	Save(p *Product) error
	GetAll() ([]Product, error)
	GetOne(id int) (*Product, error)
	// AGREGO LA FUNCION DE UPDATE
	Update(id int, p *Product) error
	Delete(id int) error
}