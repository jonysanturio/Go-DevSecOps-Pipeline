package main

import (
	"database/sql"
	"fmt"
	_ "github.com/jony/inventario/docs"
	"github.com/jony/inventario/internal/platform/postgres"
	"github.com/jony/inventario/internal/product"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger"
	"log"
	"net/http"
	"os"
	"time"
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
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/swagger/", httpSwagger.WrapHandler)

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
	
        if _, err := fmt.Fprintln(w, "Server is running and healthy"); err != nil {
            log.Printf("error escribiendo la respuesta de /health: %v", err)
}
	})

	server := &http.Server{
		Addr:               ":8080",
		Handler:            nil,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:        15 * time.Second,
		WriteTimeout:       30 * time.Second,
		IdleTimeout:        60 * time.Second,
		MaxHeaderBytes:     1 << 20,
	}
	log.Fatal(server.ListenAndServe())
}
