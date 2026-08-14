package user

import (
	"context"

	"golang.org/x/crypto/bcrypt"

	"codebasego/internal/common"
)

// Service contains user business logic.
type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, query common.PaginationQuery) ([]User, int, error) {
	return s.repo.FindAll(ctx, query)
}

func (s *Service) ListCursor(ctx context.Context, query common.CursorQuery) ([]User, common.CursorMeta, error) {
	return s.repo.FindAllCursor(ctx, query)
}

func (s *Service) GetByID(ctx context.Context, id string) (*User, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) GetByEmail(ctx context.Context, email string) (*User, error) {
	return s.repo.FindByEmail(ctx, email)
}

func (s *Service) Create(ctx context.Context, req *CreateRequest, hashedPassword string) (*User, error) {
	if req == nil {
		return nil, common.ErrBadRequest
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	entity := &User{
		Email:    req.Email,
		Password: hashedPassword,
		Name:     req.Name,
	}
	if err := s.repo.Create(ctx, entity); err != nil {
		return nil, err
	}
	return entity, nil
}

func (s *Service) Update(ctx context.Context, id string, req *UpdateRequest) (*User, error) {
	if req == nil {
		return nil, common.ErrBadRequest
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	entity, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Email != nil {
		entity.Email = *req.Email
	}
	if req.Password != nil && *req.Password != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		entity.Password = string(hashed)
	}
	if req.Name != nil {
		entity.Name = *req.Name
	}

	if err := s.repo.Update(ctx, entity); err != nil {
		return nil, err
	}
	return entity, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
