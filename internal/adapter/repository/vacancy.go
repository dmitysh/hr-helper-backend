package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"hr-helper/internal/dto_models"
	"hr-helper/internal/entity"
)

type VacancyRepository struct {
	db *pgxpool.Pool
}

func NewVacancyRepository(db *pgxpool.Pool) *VacancyRepository {
	return &VacancyRepository{
		db: db,
	}
}

func (r *VacancyRepository) Create(ctx context.Context, vacancy dto_models.CreateVacancyRequest) (uuid.UUID, error) {
	const q = `
		INSERT INTO vacancy (
id,
title,
key_requirements
)
		VALUES ($1, $2, $3)
	 RETURNING id`

	var id uuid.UUID
	err := r.db.QueryRow(ctx, q,
		vacancy.ID,
		vacancy.Title,
		vacancy.KeyRequirements,
	).Scan(&id)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("can't exec query: %w", err)
	}

	return id, nil
}

func (r *VacancyRepository) GetByID(ctx context.Context, id uuid.UUID) (entity.Vacancy, error) {
	const q = `
		SELECT
id,
title,
key_requirements,
created_at
           FROM vacancy 
          WHERE id = $1`

	var vacancy entity.Vacancy
	err := r.db.QueryRow(ctx, q, id).Scan(
		&vacancy.ID,
		&vacancy.Title,
		&vacancy.KeyRequirements,
		&vacancy.CreatedAt,
	)
	if err != nil {
		return entity.Vacancy{}, fmt.Errorf("can't exec query: %w", err)
	}

	return vacancy, nil
}
