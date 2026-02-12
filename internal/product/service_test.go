package product_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jony/inventario/internal/domain/mocks"
	"github.com/jony/inventario/internal/product"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestService_Create(t *testing.T) {
	mockRepo := new(mocks.ProductRepository)
	service := product.NewService(mockRepo)
	ctx := context.Background()

	// Caso exitoso
	t.Run("Success Create", func(t *testing.T) {
		mockRepo.On("Save", ctx, mock.AnythingOfType("*domain.Product")).Return(nil).Once()

		p, err := service.Create(ctx, "Macbook Pro", 2500.0, 10)

		assert.NoError(t, err)          
		assert.NotNil(t, p)             
		assert.Equal(t, "Macbook Pro", p.Name) 
		
		mockRepo.AssertExpectations(t)
	})

		// Caso falso
	t.Run("Database Error", func(t *testing.T) {
		mockRepo.On("Save", ctx, mock.AnythingOfType("*domain.Product")).Return(errors.New("db connection lost")).Once()

		_, err := service.Create(ctx, "Macbook Pro", 2500.0, 10)

		assert.Error(t, err) 
		assert.Equal(t, "db connection lost", err.Error())
		mockRepo.AssertExpectations(t)
	})
}