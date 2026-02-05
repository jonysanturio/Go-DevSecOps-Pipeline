package main

import "github.com/golang-jwt/jwt/v5"

// Estructura de inventario

type Product struct {
	ID		int		`json:"id"`
	Name	string	`json:"name"`
	Price	float64	`json:"price"`		
	Stock	int		`json:"stock"`
}

type Credential struct {
	Password	string	`json: password`
	Username	string	`json: username`
}

type Claims struct {
	Username	string	`json: username`
	jwt.RegisteredClaims
}

