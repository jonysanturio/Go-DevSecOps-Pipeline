package main

import (
    "encoding/json"
    "net/http"
    "strconv"

    // Asegurate que este import coincida con tu go.mod
    "github.com/jony/inventario/internal/product"
)

// ProductHandler: Maneja las peticiones web
type ProductHandler struct {
    service *product.Service
}

func NewProductHandler(service *product.Service) *ProductHandler {
    return &ProductHandler{service: service}
}

// CreateProduct crea un nuevo producto en el inventario
// @Summary      Crear un producto
// @Description  Guarda un producto con nombre, precio y stock en la base de datos
// @Tags         products
// @Accept       json
// @Produce      json
// @Param        product body product.CreateProductRequest true "Datos del producto"
// @Success      201  {object}  product.Product
// @Failure      400  {string}  string "Datos inválidos"
// @Failure      500  {string}  string "Error interno"
// @Router       /products [post]
func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
    // CAMBIO: Usamos la estructura oficial para que coincida con Swagger
    var req product.CreateProductRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "JSON Inválido", http.StatusBadRequest)
        return
    }

    // Llamar al Servicio
    p, err := h.service.Create(req.Name, req.Price, req.Stock)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(p)
}

// GetAllProducts trae la lista completa
// @Summary      Listar productos
// @Description  Obtiene todos los productos registrados
// @Tags         products
// @Produce      json
// @Success      200  {array}   product.Product
// @Failure      500  {string}  string "Error interno"
// @Router       /products [get]
func (h *ProductHandler) GetAllProducts(w http.ResponseWriter, r *http.Request) {
    products, err := h.service.GetAll()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(products)
}

// GetOneProduct obtiene un producto por ID
// @Summary      Obtener un producto
// @Description  Busca un producto específico por su ID
// @Tags         products
// @Produce      json
// @Param        id   path      int  true  "ID del Producto"
// @Success      200  {object}  product.Product
// @Failure      400  {string}  string "ID Inválido"
// @Failure      404  {string}  string "Producto no encontrado"
// @Router       /products/{id} [get]
func (h *ProductHandler) GetOneProduct(w http.ResponseWriter, r *http.Request) {
    // Extracción simple del ID (Nota: en frameworks como Gin/Chi esto es más fácil)
    idStr := r.URL.Path[len("/products/"):]
    id, err := strconv.Atoi(idStr)
    if err != nil {
        http.Error(w, "ID Inválido", http.StatusBadRequest)
        return
    }

    p, err := h.service.GetOne(id)
    if err != nil {
        http.Error(w, "Producto no encontrado", http.StatusNotFound)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(p)
}

// UpdateProduct actualiza un producto existente
// @Summary      Actualizar producto
// @Description  Actualiza los datos de un producto por su ID
// @Tags         products
// @Accept       json
// @Produce      json
// @Param        id       path      int  true  "ID del Producto"
// @Param        product  body      product.CreateProductRequest true "Nuevos datos"
// @Success      200      {object}  product.Product
// @Failure      400      {string}  string "Datos inválidos"
// @Router       /products/{id} [put]
func (h *ProductHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
    idStr := r.URL.Path[len("/products/"):]
    id, err := strconv.Atoi(idStr)
    if err != nil {
        http.Error(w, "ID Inválido", http.StatusBadRequest)
        return
    }

    var req product.CreateProductRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "JSON Inválido", http.StatusBadRequest)
        return
    }

    p, err := h.service.Update(id, req.Name, req.Price, req.Stock)
    if err != nil {
        http.Error(w, "Error actualizando producto", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(p)
}

// DeleteProduct elimina un producto
// @Summary      Eliminar producto
// @Description  Borra un producto del inventario permanentemente
// @Tags         products
// @Param        id   path      int  true  "ID del Producto"
// @Success      204  {string}  string "No Content"
// @Failure      400  {string}  string "ID Inválido"
// @Router       /products/{id} [delete]
func (h *ProductHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
    idStr := r.URL.Path[len("/products/"):]
    id, err := strconv.Atoi(idStr)
    if err != nil {
        http.Error(w, "ID Inválido", http.StatusBadRequest)
        return
    }

    err = h.service.Delete(id)
    if err != nil {
        http.Error(w, "Error eliminando producto", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusNoContent)
}