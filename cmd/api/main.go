package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"github.com/jony/inventario/internal/platform/postgres"
	"github.com/jony/inventario/internal/product"
	_ "github.com/lib/pq"
    _ "github.com/jony/inventario/docs"
    httpSwagger "github.com/swaggo/http-swagger"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)
// @title   GoStock Inventory API
// @version 1
// @description API Hexagonal para gestión de inventario con métricas y seguridad.
// @host    localhost:8080
// @BasePath    /
func main() {
    dbUser := os.Getenv("DB_USER")
    dbPass := os.Getenv("DB_PASSWORD")
    dbName := os.Getenv("DB_NAME")
    connStr := fmt.Sprintf("postgres://%s:%s@db:5432/%s?sslmode=disable", dbUser, dbPass, dbName)

    db, err := sql.Open("postgres", connStr)
    if err != nil {
        log.Fatal(err)
    }

    repo := postgres.NewRepository(db)
    service := product.NewService(repo)
    handler := NewProductHandler(service)

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


    http.HandleFunc("/swagger/", httpSwagger.WrapHandler)

    http.Handle("/metrics", promhttp.Handler())
    
    fmt.Println("SERVIDOR CORRIENDO EN PUERTO 8080")
    log.Fatal(http.ListenAndServe("0.0.0.0:8080", nil))
}