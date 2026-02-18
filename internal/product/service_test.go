package product_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jony/inventario/internal/domain"
	"github.com/jony/inventario/internal/product"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Save(ctx context.Context, p *domain.Product) error {
	args := m.Called(ctx, p)
	if p.ID == 0 {
		p.ID = 1
	}
	return args.Error(0)
}

func (m *MockRepository) GetAll(ctx context.Context) ([]domain.Product, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.Product), args.Error(1)
}

func (m *MockRepository) GetOne(ctx context.Context, id int) (*domain.Product, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Product), args.Error(1)
}

func (m *MockRepository) Update(ctx context.Context, id int, p *domain.Product) error {
	args := m.Called(ctx, id, p)
	return args.Error(0)
}

func (m *MockRepository) Delete(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestService_Create(t *testing.T) {
	mockRepo := new(MockRepository)
	service := product.NewService(mockRepo)
	ctx := context.Background()

	req := product.CreateProductRequest{
		Name:  "Macbook Pro",
		Price: 2500.0,
		Stock: 10,
	}

	t.Run("Success Create", func(t *testing.T) {
		mockRepo.On("Save", ctx, mock.AnythingOfType("*domain.Product")).Return(nil).Once()
		res, err := service.Create(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, "Macbook Pro", res.Name) 
		assert.Equal(t, 1, res.ID)               

		mockRepo.AssertExpectations(t)
	})

	t.Run("Database Error", func(t *testing.T) {
		mockRepo.On("Save", ctx, mock.AnythingOfType("*domain.Product")).Return(errors.New("db connection lost")).Once()

		_, err := service.Create(ctx, req)

		assert.Error(t, err)
		assert.Equal(t, "db connection lost", err.Error())
		mockRepo.AssertExpectations(t)
	})
}