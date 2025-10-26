package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"hr-helper/internal/dto_models"
	"hr-helper/internal/entity"
	"hr-helper/internal/inerrors"
	"hr-helper/internal/pkg/houston/loggy"
	"hr-helper/internal/service_models"
)

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

type VacancyRepository struct {
	db *pgxpool.Pool
}

func NewVacancyRepository(db *pgxpool.Pool) *VacancyRepository {
	return &VacancyRepository{
		db: db,
	}
}

func (r *VacancyRepository) CreateVacancy(ctx context.Context, vacancy dto_models.CreateVacancyRequest) (uuid.UUID, error) {
	args := []interface{}{
		vacancy.ID,
		vacancy.Title,
		vacancy.KeyRequirements,
	}

	placeholders := make([]string, 0, len(vacancy.Questions))
	for i, q := range vacancy.Questions {
		args = append(args, i+1, q.Content, q.Reference, q.TimeLimit)
		placeholders = append(placeholders, fmt.Sprintf("($%d::int, $%d::text, $%d::text, $%d::int)",
			len(args)-3, len(args)-2, len(args)-1, len(args)))
	}

	q := fmt.Sprintf(`
        WITH vacancy_insert AS (
            INSERT INTO vacancy (id, title, key_requirements)
            VALUES ($1, $2, $3)
          RETURNING id
        ),
        questions_insert AS (
            INSERT INTO question (vacancy_id, position, content, reference, time_limit)
            SELECT $1, position, content, reference, time_limit
            FROM (VALUES %s) AS t(position, content, reference, time_limit)
        )
		SELECT id FROM vacancy_insert
    `, strings.Join(placeholders, ","))

	loggy.Info(q)
	loggy.Info(args)

	var id uuid.UUID
	err := r.db.QueryRow(ctx, q, args...).Scan(&id)
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

func (r *VacancyRepository) CreateQuestions(ctx context.Context, vacancyID uuid.UUID, questions []dto_models.CreateQuestionRequest) error {
	insertBuilder := psql.Insert("question").
		Columns(
			"vacancy_id",
			"content",
			"reference",
			"time_limit",
			"position",
		)

	for i, question := range questions {
		insertBuilder.Values(vacancyID, question.Content, question.Reference, question.TimeLimit, i+1)
	}

	q, args, err := insertBuilder.ToSql()
	if err != nil {
		return fmt.Errorf("can't build query: %w", err)
	}

	_, err = r.db.Exec(ctx, q, args...)
	if err != nil {
		return fmt.Errorf("can't exec query: %w", err)
	}

	return nil
}

func (r *VacancyRepository) GetQuestionByID(ctx context.Context, id int64) (entity.Question, error) {
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

func (r *VacancyRepository) CreateAnswer(ctx context.Context, answer service_models.ScoredAnswer) (int64, error) {
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

func (r *VacancyRepository) GetQuestionsByVacancyID(ctx context.Context, vacancyID uuid.UUID) ([]entity.Question, error) {
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
	  ORDER BY position
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
