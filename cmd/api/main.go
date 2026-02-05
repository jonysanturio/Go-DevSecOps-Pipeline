package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	// Importamos nuestros paquetes internos
	"github.com/jony/inventario/internal/platform/postgres"
	"github.com/jony/inventario/internal/product"
	_ "github.com/lib/pq"
)

func main() {
    // CORRECCIÓN 2: Es Getenv (Get Environment)
    dbUser := os.Getenv("DB_USER")
    dbPass := os.Getenv("DB_PASSWORD")
    dbName := os.Getenv("DB_NAME")
    connStr := fmt.Sprintf("postgres://%s:%s@db:5432/%s?sslmode=disable", dbUser, dbPass, dbName)

    // CORRECCIÓN 3: Es sql.Open (del paquete sql)
    db, err := sql.Open("postgres", connStr)
    if err != nil {
        log.Fatal(err)
    }

    repo := postgres.NewRepository(db)
    service := product.NewService(repo)
    handler := NewProductHandler(service)

    // CORRECCIÓN 4: Es HandleFunc (función), no HandlerFunc (tipo)
    http.HandleFunc("/products", func(w http.ResponseWriter, r *http.Request) {
        if r.Method == http.MethodPost {
            handler.CreateProduct(w, r)
        } else {
            handler.GetAllProducts(w, r)
        }
    })
    http.HandleFunc("/products/", func(w http.ResponseWriter, r *http.Request) {
        switch r.Method {
        case http.MethodGet:
            handler.GetOneProduct(w, r)
        case http.MethodPut:
            handler.UpdateProduct(w, r)
        case http.MethodDelete:
            handler.DeleteProduct(w, r)
        }
    })

    fmt.Println("SERVIDOR CORRIENDO EN PUERTO 8080")
    log.Fatal(http.ListenAndServe("0.0.0.0:8080", nil))
}