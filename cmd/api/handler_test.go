package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jony/inventario/internal/product"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockService struct {
	mock.Mock
}

func (m *MockService) Create(ctx context.Context, req product.CreateProductRequest) (*product.ProductResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*product.ProductResponse), args.Error(1)
}

func (m *MockService) GetAll(ctx context.Context) ([]product.ProductResponse, error) {
	return nil, nil
}

func (m *MockService) GetOne(ctx context.Context, id int) (*product.ProductResponse, error) {
	return nil, nil
}

func (m *MockService) Update(ctx context.Context, id int, req product.CreateProductRequest) (*product.ProductResponse, error) {
	return nil, nil
}

func (m *MockService) Delete(ctx context.Context, id int) error {
	return nil
}

func TestCreateProduct_Handler(t *testing.T) {
	mockSvr := new(MockService)
	handler := NewProductHandler(mockSvr)

	productReq := product.CreateProductRequest{
		Name:  "Mouse",
		Price: 50.0,
		Stock: 10,
	}
	body, _ := json.Marshal(productReq)

	expectedResponse := &product.ProductResponse{
		ID:    1,
		Name:  "Mouse",
		Price: 50.0,
		Stock: 10,
	}

	mockSvr.On("Create", mock.Anything, productReq).Return(expectedResponse, nil)

	req, _ := http.NewRequest("POST", "/products", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()
	
	handler.CreateProduct(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
	mockSvr.AssertExpectations(t)
}