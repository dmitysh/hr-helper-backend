package repository

import (
	"context"
	"encoding/json"
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
		INSERT INTO answer (
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

func (r *VacancyRepository) GetAnswers(ctx context.Context, candidateID int64, vacancyID uuid.UUID) ([]entity.Answer, error) {
	const q = `
		SELECT
answer.id,
answer.candidate_id,
answer.question_id,
answer.content,
answer.score,
answer.time_taken,
answer.created_at
          FROM answer 
          JOIN question 
            ON answer.question_id = question.id
		 WHERE candidate_id = $1
		   AND question.vacancy_id = $2
         `

	rows, err := r.db.Query(ctx, q, candidateID, vacancyID)
	if err != nil {
		return nil, fmt.Errorf("can't query: %w", err)
	}

	answers, err := pgx.CollectRows(rows, pgx.RowToStructByName[entity.Answer])
	if err != nil {
		return nil, fmt.Errorf("can't collect rows: %w", err)
	}

	return answers, nil
}

func (r *VacancyRepository) UpdateInterviewResult(ctx context.Context, candidateID int64, vacancyID uuid.UUID, result service_models.InterviewResult) error {
	const upsertMetaQuery = `
		INSERT INTO candidate_vacancy_meta (
candidate_id,
vacancy_id,
interview_score,		                                    
status,      
updated_at
)
		VALUES ($1, $2, $3, $4, now())
   ON CONFLICT (candidate_id, vacancy_id)
	 DO UPDATE
		   SET 
interview_score = EXCLUDED.interview_score,
status          = EXCLUDED.status,
updated_at      = now();`

	_, err := r.db.Exec(ctx, upsertMetaQuery,
		candidateID,
		vacancyID,
		result.Score,
		result.Status,
	)
	if err != nil {
		return fmt.Errorf("can't exec query: %w", err)
	}

	return nil
}

func (r *VacancyRepository) GetVacanciesWithQuestions(ctx context.Context) ([]entity.VacancyWithQuestion, error) {
	const q = `
		SELECT
v.id,
v.title,
v.key_requirements,
v.created_at,
COALESCE(
json_agg(
	jsonb_build_object(
		'id', q.id,
		'vacancy_id', q.vacancy_id,
		'content', q.content,
		'reference', q.reference,
		'time_limit', q.time_limit,
		'position', q.position,
		'created_at', q.created_at
	) ORDER BY q.position),
	'[]'::json
) AS questions
     FROM vacancy v
LEFT JOIN question q ON q.vacancy_id = v.id
 GROUP BY v.id, v.title, v.key_requirements, v.created_at
 ORDER BY v.created_at DESC`

	var vacancies []entity.VacancyWithQuestion
	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("can't exec query: %w", err)
	}

	for rows.Next() {
		var v entity.VacancyWithQuestion
		var questionsJSON []byte
		err = rows.Scan(&v.ID, &v.Title, &v.KeyRequirements, &v.CreatedAt, &questionsJSON)
		if err != nil {
			return nil, fmt.Errorf("can't scan vacancy: %w", err)
		}
		err = json.Unmarshal(questionsJSON, &v.Questions)
		if err != nil {
			return nil, fmt.Errorf("can't unmarshal questions: %w", err)
		}

		vacancies = append(vacancies, v)
	}

	return vacancies, nil
}

func (r *VacancyRepository) GetVacancyWithQuestions(ctx context.Context, vacancyID uuid.UUID) (entity.VacancyWithQuestion, error) {
	const q = `
		SELECT
v.id,
v.title,
v.key_requirements,
v.created_at,
COALESCE(
json_agg(
	jsonb_build_object(
		'id', q.id,
		'vacancy_id', q.vacancy_id,
		'content', q.content,
		'reference', q.reference,
		'time_limit', q.time_limit,
		'position', q.position,
		'created_at', q.created_at
	) ORDER BY q.position),
	'[]'::json
) AS questions
     FROM vacancy v
LEFT JOIN question q ON q.vacancy_id = v.id
  WHERE v.id = $1
 GROUP BY v.id, v.title, v.key_requirements, v.created_at
 ORDER BY v.created_at DESC`

	row := r.db.QueryRow(ctx, q, vacancyID)

	var v entity.VacancyWithQuestion
	var questionsJSON []byte
	err := row.Scan(&v.ID, &v.Title, &v.KeyRequirements, &v.CreatedAt, &questionsJSON)
	if errors.Is(err, pgx.ErrNoRows) {
		return entity.VacancyWithQuestion{}, inerrors.ErrNotFound
	}

	if err != nil {
		return entity.VacancyWithQuestion{}, fmt.Errorf("can't scan vacancy: %w", err)
	}
	err = json.Unmarshal(questionsJSON, &v.Questions)
	if err != nil {
		return entity.VacancyWithQuestion{}, fmt.Errorf("can't unmarshal questions: %w", err)
	}

	return v, nil
}
