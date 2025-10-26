package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"hr-helper/internal/dto_models"
	"hr-helper/internal/entity"
	"hr-helper/internal/inerrors"
	"hr-helper/internal/service_models"
)

type CandidateRepository struct {
	db *pgxpool.Pool
}

func NewCandidateRepository(db *pgxpool.Pool) *CandidateRepository {
	return &CandidateRepository{
		db: db,
	}
}

func (r *CandidateRepository) Create(ctx context.Context, candidate dto_models.CreateCandidateRequest) (int64, error) {
	const q = `
		INSERT INTO candidate (
telegram_id,
full_name,
phone,
city
)
		VALUES ($1, $2, $3, $4)
	 RETURNING id`

	var id int64
	err := r.db.QueryRow(ctx, q,
		candidate.TelegramID,
		candidate.FullName,
		candidate.Phone,
		candidate.City,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("can't exec query: %w", err)
	}

	return id, nil
}

func (r *CandidateRepository) GetByTelegramID(ctx context.Context, telegramID int64) (entity.Candidate, error) {
	const q = `
		SELECT 
id,
telegram_id,
full_name,
phone,
city,
created_at
		  FROM candidate
		 WHERE telegram_id = $1`

	var candidate entity.Candidate
	err := r.db.QueryRow(ctx, q, telegramID).Scan(
		&candidate.ID,
		&candidate.TelegramID,
		&candidate.FullName,
		&candidate.Phone,
		&candidate.City,
		&candidate.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return entity.Candidate{}, inerrors.ErrNotFound
	}
	if err != nil {
		return entity.Candidate{}, fmt.Errorf("can't exec query: %w", err)
	}

	return candidate, nil
}

func (r *CandidateRepository) UpdateScreeningResult(ctx context.Context, candidateID int64, vacancyID uuid.UUID, result service_models.ResumeScreeningResult) error {
	const q = `
		INSERT INTO resume_screening (
candidate_id,
vacancy_id,
score,
feedback,
updated_at
)
		VALUES ($1, $2, $3, $4, now())
   ON CONFLICT (candidate_id, vacancy_id)
	 DO UPDATE
		   SET 
score    = EXCLUDED.score,
feedback = EXCLUDED.feedback,
updated_at = now();`

	_, err := r.db.Exec(ctx, q,
		candidateID,
		vacancyID,
		result.Score,
		result.Feedback,
	)
	if err != nil {
		return fmt.Errorf("can't exec query: %w", err)
	}

	return nil
}

func (r *CandidateRepository) GetResumeScreening(ctx context.Context, candidateID int64, vacancyID uuid.UUID) (entity.ResumeScreening, error) {
	const q = `
		SELECT 
id,
candidate_id,
vacancy_id,
score,
feedback,
created_at,
updated_at
		  FROM resume_screening
		 WHERE candidate_id = $1 
		   AND vacancy_id = $2`

	var resumeScreening entity.ResumeScreening
	err := r.db.QueryRow(ctx, q, candidateID, vacancyID).Scan(
		&resumeScreening.ID,
		&resumeScreening.CandidateID,
		&resumeScreening.VacancyID,
		&resumeScreening.Score,
		&resumeScreening.Feedback,
		&resumeScreening.CreatedAt,
		&resumeScreening.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return entity.ResumeScreening{}, inerrors.ErrNotFound
	}
	if err != nil {
		return entity.ResumeScreening{}, fmt.Errorf("can't exec query: %w", err)
	}

	return resumeScreening, nil
}
