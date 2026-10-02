package main

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/jony/inventario/internal/domain"
	"github.com/jony/inventario/internal/kit/httphelper"
	"github.com/jony/inventario/internal/product"
)

type ProductService interface {
	Create(ctx context.Context, req product.CreateProductRequest) (*product.ProductResponse, error)
	GetAll(ctx context.Context) ([]product.ProductResponse, error)
	GetOne(ctx context.Context, id int) (*product.ProductResponse, error)
	Update(ctx context.Context, id int, req product.CreateProductRequest) (*product.ProductResponse, error)
	Delete(ctx context.Context, id int) error
}

type ProductHandler struct {
	service ProductService 
}

func NewProductHandler(service ProductService) *ProductHandler {
	return &ProductHandler{service: service}
}

func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	req, err := httphelper.Decode[product.CreateProductRequest](r)
	if err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	p, err := h.service.Create(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	httphelper.Encode(w, http.StatusCreated, p)
}

func (h *ProductHandler) GetAllProducts(w http.ResponseWriter, r *http.Request) {
	products, err := h.service.GetAll(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	httphelper.Encode(w, http.StatusOK, products)
}

func (h *ProductHandler) GetOneProduct(w http.ResponseWriter, r *http.Request) {
	id, err := parseProductID(r.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if id <= 0{
		http.Error(w, "ID debe ser mayor que cero", http.StatusBadRequest)
		return
	}

	p, err := h.service.GetOne(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrProductNotFound) {
			http.Error(w, "Producto no encontrado", http.StatusNotFound)
			return
		}
		http.Error(w, "Error buscando producto", http.StatusInternalServerError)
		return
	}
	httphelper.Encode(w, http.StatusOK, p)
}

func (h *ProductHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/products/"):]
	productID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID Inválido", http.StatusBadRequest)
		return
	}
	if productID <= 0 {
		http.Error(w, "ID debe ser mayor a cero", http.StatusBadRequest)
		return
	}

	updateReq, err := httphelper.Decode[product.CreateProductRequest](r)
	if err != nil {
		http.Error(w, "JSON Inválido", http.StatusBadRequest)
		return
	}

	if err := updateReq.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}


	updateProduct, err := h.service.Update(r.Context(), productID, updateReq)
	if err != nil {
		if errors.Is(err, domain.ErrProductNotFound) {
			http.Error(w, "Producto no encontrado", http.StatusNotFound)
			return
		}
		http.Error(w, "Error actualizando producto", http.StatusInternalServerError)
		return
	}
	httphelper.Encode(w, http.StatusOK, updateProduct)
}

func (h *ProductHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/products/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID Inválido", http.StatusBadRequest)
		return
	}
	if id <= 0 {
		http.Error(w, "ID debe ser mayor a cero", http.StatusBadRequest)
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		if errors.Is(err, domain.ErrProductNotFound) {
			http.Error(w, "Producto no encontrado", http.StatusNotFound)
			return
		}
		http.Error(w, "Error eliminando producto", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func parseProductID(path string) (int, error){
	idStr := path[len("/products/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, errors.New("ID Inválido")
	}
	return id, nil 
}

