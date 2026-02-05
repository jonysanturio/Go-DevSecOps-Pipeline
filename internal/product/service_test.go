package product_test

import (
	"testing"

	"github.com/jony/inventario/internal/domain"
	"github.com/jony/inventario/internal/product"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MOCK (Repo Mentiroso)
// Simula ser un repositorio real

type MockRepository struct{
	mock.Mock
}

func (m *MockRepository)Save(p *domain.Product)error {
	args := m.Called(p)
	return args.Error(0)
}

func (m *MockRepository)GetAll()([]domain.Product, error){
	args := m.Called()
	return args.Get(0).([]domain.Product), args.Error(1)
}

func (m *MockRepository)GetOne(id int)(*domain.Product, error){
	return nil, nil
}
func (m *MockRepository)Update(id int, p*domain.Product)error{
	return nil
}

func (m *MockRepository)Delete(id int) error{
	return nil
}

// Testing REAL

func TestCreate_Success(t *testing.T){
	repoMock := new(MockRepository)
	service := product.NewService(repoMock)

	// SI LLAMAN A SAVE, DEVUELVE NIL
	repoMock.On("Save", mock.Anything).Return(nil)
	// Ejecución ACT
	p, err := service.Create("Coca-Cola", 10.5, 100)
	// VERIFICAR
	assert.NoError(t, err)
	assert.NotNil(t, p)
	assert.Equal(t, "Coca-Cola", p.Name)
	//VERIFICAR EL SERVICIO LLAMADO REPOSITORIO
	repoMock.AssertExpectations(t)
}

func TestCreate_InvalidPrice(t *testing.T){
	repoMock := new(MockRepository)
	service := product.NewService(repoMock)

	_, err := service.Create("Coca-Cola", -50.0, 100)

	assert.Error(t, err)
	assert.Equal(t, domain.ErrInvalidPrice, err)

	repoMock.AssertNotCalled(t, "Save")
}
