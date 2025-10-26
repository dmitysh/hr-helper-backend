package vacancy

import (
	"context"

	"github.com/google/uuid"

	"hr-helper/internal/dto_models"
)

type Storage interface {
	Create(ctx context.Context, vacancy dto_models.CreateVacancyRequest) (uuid.UUID, error)
}

type Service struct {
	store Storage
}

func NewService(store Storage) *Service {
	return &Service{
		store: store,
	}
}

func (s *Service) Create(ctx context.Context, vacancy dto_models.CreateVacancyRequest) (uuid.UUID, error) {
	return s.store.Create(ctx, vacancy)
}
