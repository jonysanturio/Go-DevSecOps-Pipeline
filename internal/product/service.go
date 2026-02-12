package product

import (
	"context"

	"github.com/jony/inventario/internal/domain"
)

type Service struct {
	repo domain.ProductRepository
}

func NewService(repo domain.ProductRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, req CreateProductRequest) (*ProductResponse, error) {
	p := &domain.Product{
		Name:  req.Name,
		Price: req.Price,
		Stock: req.Stock,
	}

	if err := p.Validate(); err != nil {
		return nil, err
	}

	if err := s.repo.Save(ctx, p); err != nil {
		return nil, err
	}

	res := ToResponse(p)
	return &res, nil
}

func (s *Service) GetAll(ctx context.Context) ([]ProductResponse, error) {
	products, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	return ToResponseList(products), nil
}

func (s *Service) GetOne(ctx context.Context, id int) (*ProductResponse, error) {
	p, err := s.repo.GetOne(ctx, id)
	if err != nil {
		return nil, err
	}
	res := ToResponse(p)
	return &res, nil
}

func (s *Service) Update(ctx context.Context, id int, req CreateProductRequest) (*ProductResponse, error) {
	p := &domain.Product{
		ID:    id,
		Name:  req.Name,
		Price: req.Price,
		Stock: req.Stock,
	}

	if err := p.Validate(); err != nil {
		return nil, err
	}

	if err := s.repo.Update(ctx, id, p); err != nil {
		return nil, err
	}

	res := ToResponse(p)
	return &res, nil
}

func (s *Service) Delete(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}