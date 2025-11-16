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
telegram_username,
full_name,
phone,
city
)
		VALUES ($1, $2, $3, $4, $5)
	 RETURNING id`

	var id int64
	err := r.db.QueryRow(ctx, q,
		candidate.TelegramID,
		candidate.TelegramUsername,
		candidate.FullName,
		candidate.Phone,
		candidate.City,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("can't exec query: %w", err)
	}

	return id, nil
}

func (r *CandidateRepository) Delete(ctx context.Context, candidateID int64) error {
	const q = `
		DELETE FROM candidate
	     WHERE id = $1`

	_, err := r.db.Exec(ctx, q, candidateID)
	if err != nil {
		return fmt.Errorf("can't exec query: %w", err)
	}

	return nil
}

func (r *CandidateRepository) GetByTelegramID(ctx context.Context, telegramID int64) (entity.Candidate, error) {
	const q = `
		SELECT 
id,
telegram_id,
telegram_username,
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
		&candidate.TelegramUsername,
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

func (r *CandidateRepository) UpdateScreeningResult(ctx context.Context, candidateID int64, vacancyID uuid.UUID, result service_models.ResumeScreeningResultWithStatus) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("can't begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	const upsertResumeScreeningQuery = `
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

	_, err = tx.Exec(ctx, upsertResumeScreeningQuery,
		candidateID,
		vacancyID,
		result.Score,
		result.Feedback,
	)
	if err != nil {
		return fmt.Errorf("can't exec query: %w", err)
	}

	const upsertMetaQuery = `
		INSERT INTO candidate_vacancy_meta (
candidate_id,
vacancy_id,  
status,      
updated_at 
)
		VALUES ($1, $2, $3, now())
   ON CONFLICT (candidate_id, vacancy_id)
	 DO UPDATE
		   SET 
status    = EXCLUDED.status,
updated_at = now();`

	_, err = tx.Exec(ctx, upsertMetaQuery,
		candidateID,
		vacancyID,
		result.Status,
	)
	if err != nil {
		return fmt.Errorf("can't exec query: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("can't commit tx: %w", err)
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

func (r *CandidateRepository) GetMeta(ctx context.Context, candidateID int64, vacancyID uuid.UUID) (entity.Meta, error) {
	const q = `
		SELECT 
candidate_id,
vacancy_id,  
interview_score,
status,      
updated_at
		  FROM candidate_vacancy_meta
		 WHERE candidate_id = $1 
		   AND vacancy_id = $2`

	var meta entity.Meta
	err := r.db.QueryRow(ctx, q, candidateID, vacancyID).Scan(
		&meta.CandidateID,
		&meta.VacancyID,
		&meta.InterviewScore,
		&meta.Status,
		&meta.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return entity.Meta{}, inerrors.ErrNotFound
	}
	if err != nil {
		return entity.Meta{}, fmt.Errorf("can't exec query: %w", err)
	}

	return meta, nil
}

func (r *CandidateRepository) GetCandidateVacancyInfos(ctx context.Context) ([]entity.CandidateVacancyInfo, error) {
	const q = `
		SELECT 
    c.id AS candidate_id,
    c.telegram_id,
    c.full_name,
    c.phone,
    c.city,
    c.created_at AS candidate_created_at,

    v.id AS vacancy_id,
    v.title,
    v.key_requirements,
    v.created_at AS vacancy_created_at,

    m.candidate_id AS meta_candidate_id,
    m.vacancy_id AS meta_vacancy_id,
    m.interview_score,
    m.status,
    m.is_archived,
    m.updated_at,
    
    rs.id,
    rs.score,
    rs.feedback,
    rs.created_at,
    rs.updated_at

FROM candidate c
JOIN candidate_vacancy_meta m ON m.candidate_id = c.id
JOIN vacancy v ON v.id = m.vacancy_id
JOIN resume_screening rs ON rs.candidate_id = c.id AND rs.vacancy_id = v.id`

	var infos []entity.CandidateVacancyInfo
	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("can't exec query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var info entity.CandidateVacancyInfo
		var keyRequirements []string

		err = rows.Scan(
			&info.Candidate.ID,
			&info.Candidate.TelegramID,
			&info.Candidate.FullName,
			&info.Candidate.Phone,
			&info.Candidate.City,
			&info.Candidate.CreatedAt,

			&info.Vacancy.ID,
			&info.Vacancy.Title,
			&keyRequirements,
			&info.Vacancy.CreatedAt,

			&info.Meta.CandidateID,
			&info.Meta.VacancyID,
			&info.Meta.InterviewScore,
			&info.Meta.Status,
			&info.Meta.IsArchived,
			&info.Meta.UpdatedAt,

			&info.ResumeScreening.ID,
			&info.ResumeScreening.Score,
			&info.ResumeScreening.Feedback,
			&info.ResumeScreening.CreatedAt,
			&info.ResumeScreening.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("can't scan row: %w", err)
		}
		info.Vacancy.KeyRequirements = keyRequirements

		infos = append(infos, info)
	}

	return infos, nil
}

func (r *CandidateRepository) GetCandidateVacancyInfo(ctx context.Context, candidateID int64, vacancyID uuid.UUID) (entity.CandidateVacancyInfo, error) {
	const q = `
		SELECT 
    c.id AS candidate_id,
    c.telegram_id,
    c.telegram_username,
    c.full_name,
    c.phone,
    c.city,
    c.created_at AS candidate_created_at,

    v.id AS vacancy_id,
    v.title,
    v.key_requirements,
    v.created_at AS vacancy_created_at,

    m.candidate_id AS meta_candidate_id,
    m.vacancy_id AS meta_vacancy_id,
    m.interview_score,
    m.status,
    m.is_archived,
    m.updated_at,
    
    rs.id,
    rs.score,
    rs.feedback,
    rs.created_at,
    rs.updated_at

 FROM candidate c
 JOIN candidate_vacancy_meta m ON m.candidate_id = c.id
 JOIN vacancy v ON v.id = m.vacancy_id
 JOIN resume_screening rs ON rs.candidate_id = c.id AND rs.vacancy_id = v.id
WHERE c.id = $1 AND v.id = $2`

	var info entity.CandidateVacancyInfo
	var keyRequirements []string

	row := r.db.QueryRow(ctx, q, candidateID, vacancyID)

	err := row.Scan(
		&info.Candidate.ID,
		&info.Candidate.TelegramID,
		&info.Candidate.TelegramUsername,
		&info.Candidate.FullName,
		&info.Candidate.Phone,
		&info.Candidate.City,
		&info.Candidate.CreatedAt,

		&info.Vacancy.ID,
		&info.Vacancy.Title,
		&keyRequirements,
		&info.Vacancy.CreatedAt,

		&info.Meta.CandidateID,
		&info.Meta.VacancyID,
		&info.Meta.InterviewScore,
		&info.Meta.Status,
		&info.Meta.IsArchived,
		&info.Meta.UpdatedAt,

		&info.ResumeScreening.ID,
		&info.ResumeScreening.Score,
		&info.ResumeScreening.Feedback,
		&info.ResumeScreening.CreatedAt,
		&info.ResumeScreening.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return entity.CandidateVacancyInfo{}, inerrors.ErrNotFound
	}
	if err != nil {
		return entity.CandidateVacancyInfo{}, fmt.Errorf("can't scan row: %w", err)
	}
	info.Vacancy.KeyRequirements = keyRequirements

	return info, nil
}

func (r *CandidateRepository) GetCandidateAnswers(ctx context.Context, candidateID int64, vacancyID uuid.UUID) ([]entity.CandidateQuestionAnswer, error) {
	const q = `
		SELECT
    q.id,
    q.vacancy_id,
    q.content,
    q.reference,

    a.id,
    a.candidate_id,
    a.question_id,
    a.content,
    a.score,
    a.time_taken,
    a.created_at

FROM candidate c
         JOIN candidate_vacancy_meta m ON m.candidate_id = c.id
         JOIN question q ON q.vacancy_id = m.vacancy_id
         JOIN answer a ON a.candidate_id =c.id AND q.id = a.question_id
WHERE c.id = $1 AND m.vacancy_id = $2
ORDER BY q.position`

	var questionAnswers []entity.CandidateQuestionAnswer

	rows, err := r.db.Query(ctx, q, candidateID, vacancyID)
	if err != nil {
		return nil, fmt.Errorf("can't exec query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var questionAnswer entity.CandidateQuestionAnswer

		err = rows.Scan(
			&questionAnswer.Question.ID,
			&questionAnswer.Question.VacancyID,
			&questionAnswer.Question.Content,
			&questionAnswer.Question.Reference,

			&questionAnswer.Answer.ID,
			&questionAnswer.Answer.CandidateID,
			&questionAnswer.Answer.QuestionID,
			&questionAnswer.Answer.Content,
			&questionAnswer.Answer.Score,
			&questionAnswer.Answer.TimeTaken,
			&questionAnswer.Answer.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("can't scan row: %w", err)
		}

		questionAnswers = append(questionAnswers, questionAnswer)
	}

	return questionAnswers, nil
}
