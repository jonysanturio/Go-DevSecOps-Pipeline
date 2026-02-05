package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/jony/inventario/internal/product"
)

// ProductHandler: Maneja las peticiones web
type ProductHandler struct {
	service *product.Service
}

func NewProductHandler(service *product.Service) *ProductHandler{
	return  &ProductHandler{service: service}
}

// CreateProduct: POST /products

func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request){
	// 1. Decodificar JSON (Input)
	var req struct {
		Name	string	`json:"name"`
		Price	float64	`json:"price"`
		Stock	int		`json:"stock"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil{
		http.Error(w,"Json Invalidado", http.StatusBadRequest)
		return
	}
	// 2. Llamar al Servicio (Proceso)
	p, err := h.service.Create(req.Name, req.Price, req.Stock)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// 3. Responder JSON (Output)
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(p)
}

// GetAllProducts: GET /products
func (h *ProductHandler) GetAllProducts(w http.ResponseWriter, r *http.Request){
	products, err := h.service.GetAll()
	if err != nil{
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-type", "application/json")
	json.NewEncoder(w).Encode(products)
}

// GetOne: Post/ products

func (h *ProductHandler)GetOneProduct(w http.ResponseWriter, r *http.Request){
	idStr := r.URL.Path[len("/products/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil{
		http.Error(w, "ID Invalido", http.StatusBadRequest)
		return
	}
	p, err := h.service.GetOne(id)
	if err != nil{
		http.Error(w, "Error no se encontro producto", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

func (h *ProductHandler)UpdateProduct(w http.ResponseWriter, r *http.Request){
	idStr := r.URL.Path[len("/products/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil{
		http.Error(w, "ID Invalido", http.StatusBadRequest)
		return 
	}
	// DECODIFICAR EL JSON EN EL SCANNER
	var req struct{
		Name	string	`json:"name"`
		Price	float64	`json:"price"`
		Stock	int		`json:"stock"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil{
		http.Error(w, "JSON INVALIDO", http.StatusBadRequest)
		return
	}

	// LLAMAR AL SERVICIO
	p, err := h.service.Update(id, req.Name, req.Price, req.Stock)
	if err != nil {
		http.Error(w, "ERROR ACTUALIZANDO PRODUCTO", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

func (h *ProductHandler)DeleteProduct(w http.ResponseWriter, r *http.Request){
	idStr := r.URL.Path[len("/products/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil{
		http.Error(w, "ID INVALIDO", http.StatusBadRequest)
		return
	}
	// LLAMO AL SERVIDOR
	err = h.service.Delete(id)
	if err != nil{
		http.Error(w, "ERROR ELIMINANDO PRODUCTO", http.StatusInternalServerError)
		return
	}
	// RESPUESTA DEL 204
	w.WriteHeader(http.StatusNoContent)
}
