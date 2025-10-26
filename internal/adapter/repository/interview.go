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

type InterviewRepository struct {
	db *pgxpool.Pool
}

func NewInterviewRepository(db *pgxpool.Pool) *InterviewRepository {
	return &InterviewRepository{
		db: db,
	}
}

func (r *InterviewRepository) CreateQuestion(ctx context.Context, question dto_models.CreateQuestionRequest) (int64, error) {
	const q = `
		INSERT INTO question (
vacancy_id,
content,
reference,
time_limit,
position
)
		VALUES ($1, $2, $3, $4, $5)
	 RETURNING id`

	var id int64
	err := r.db.QueryRow(ctx, q,
		question.VacancyID,
		question.Content,
		question.Reference,
		question.Position,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("can't exec query: %w", err)
	}

	return id, nil
}

func (r *InterviewRepository) GetQuestionByID(ctx context.Context, id int64) (entity.Question, error) {
	const q = `
		SELECT
id,
vacancy_id,
content,
reference,
time_limit,
position,
created_at
          FROM question
          WHERE id = $1`

	var question entity.Question
	err := r.db.QueryRow(ctx, q, id).Scan(
		&question.ID,
		&question.VacancyID,
		&question.Content,
		&question.Reference,
		&question.TimeLimit,
		&question.Position,
		&question.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return entity.Question{}, inerrors.ErrNotFound
	}
	if err != nil {
		return entity.Question{}, fmt.Errorf("can't exec query: %w", err)
	}

	return question, nil
}

func (r *InterviewRepository) CreateAnswer(ctx context.Context, answer service_models.ScoredAnswer) (int64, error) {
	const q = `
		INSERT INTO question (
candidate_id,
question_id,
content,
score,
time_taken
)
		VALUES ($1, $2, $3, $4, $5)
	 RETURNING id`

	var id int64
	err := r.db.QueryRow(ctx, q,
		answer.CandidateID,
		answer.QuestionID,
		answer.Content,
		answer.Score,
		answer.TimeTaken,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("can't exec query: %w", err)
	}

	return id, nil
}

func (r *InterviewRepository) GetQuestionsByVacancyID(ctx context.Context, vacancyID uuid.UUID) ([]entity.Question, error) {
	const q = `
		SELECT
id,
vacancy_id,
content,
reference,
time_limit,
"position",
created_at
          FROM question
		 WHERE vacancy_id = $1
         `

	rows, err := r.db.Query(ctx, q, vacancyID)
	if err != nil {
		return nil, fmt.Errorf("can't query: %w", err)
	}

	questions, err := pgx.CollectRows(rows, pgx.RowToStructByName[entity.Question])
	if err != nil {
		return nil, fmt.Errorf("can't collect rows: %w", err)
	}

	return questions, nil
}
