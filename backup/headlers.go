package main

import (
	"encoding/json"
	"database/sql"
	"fmt"
	"log"
	"net/http"
)

func createTable(db *sql.DB){
		query := `
		CREATE TABLE IF NOT EXISTS products (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			price DECIMAL(10, 2) NOT NULL DEFAULT 0.00,
			stock INT NOT NULL DEFAULT 0
		);`

		_, err := db.Exec(query)
		if err != nil{
			log.Fatal("Error creado la tabla", err)
		}
		fmt.Println("Tabla Product creada y se puede usar")
	}

	// 7. Función para CREAR un producto (POST)
	func createProductHandler(w http.ResponseWriter, r*http.Request, db *sql.DB){
		// Seguridad:  Solo permitir metodo POST
		if r.Method != http.MethodPost {
			http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
			return
		}
		var p Product
		// Decodificamos el JSON que nos envian 
		err := json.NewDecoder(r.Body).Decode(&p)
		if err != nil{
			http.Error(w, "Error leyendo el JSON", http.StatusBadRequest)
			return
		}
		// Query SQL Segura (Con parámetros $1, $2, $3 para evitar inyecciones)
		query := `INSERT INTO products (name, price, stock) VALUES ($1, $2, $3) RETURNING id`
		// EJECUTAR Y GUARDAR EL ID
		err = db.QueryRow(query, p.Name, p.Price, p.Stock).Scan(&p.ID)
		if err != nil{
			http.Error(w,"Error guardando la BD: "+err.Error(), http.StatusInternalServerError)
			return
		}
		// Responder a los productos
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(p)
	}	

	// 8. FUNCION DE LISTAR productos (GET)
	func getProductsHandler(w http.ResponseWriter, r*http.Request, db *sql.DB){
		if r.Method != http.MethodGet {
			http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
			return
		}
		rows, err := db.Query("SELECT id, name, price, stock FROM products")
		if err != nil{
			http.Error(w, "Error consultando BD", http.StatusInternalServerError)
			return
		}
		defer rows.Close() //Cerrar filas para no ahogar la BD
		var products []Product
		for rows.Next(){
			var p Product
			if err := rows.Scan(&p.ID, &p.Name, &p.Price, &p.Stock); err != nil{
				continue
			}
			products = append(products, p)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(products)
	}