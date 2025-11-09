package dto_models

import (
	"time"

	"github.com/google/uuid"
)

type CreateVacancyRequest struct {
	ID              uuid.UUID               `json:"id"`
	Title           string                  `json:"title"`
	KeyRequirements []string                `json:"key_requirements"`
	Questions       []CreateQuestionRequest `json:"questions"`
}
type CreateQuestionRequest struct {
	Content   string `json:"content"`
	Reference string `json:"reference"`
	TimeLimit int    `json:"time_limit"`
}

type CreateAnswerRequest struct {
	CandidateID int64  `json:"candidate_id"`
	QuestionID  int64  `json:"question_id"`
	Content     string `json:"content"`
	TimeTaken   int    `json:"time_taken"`
}

type GetQuestionResponse struct {
	ID        int64     `json:"id"`
	VacancyID uuid.UUID `json:"vacancy_id"`
	Content   string    `json:"content"`
	Reference string    `json:"reference"`
	TimeLimit int       `json:"time_limit"`
	Position  int       `json:"position"`
	CreatedAt time.Time `json:"created_at"`
}

type GetVacancyWithQuestionsResponse struct {
	ID              uuid.UUID             `json:"id"`
	Title           string                `json:"title"`
	KeyRequirements []string              `json:"key_requirements"`
	Questions       []GetQuestionResponse `json:"questions"`
	CreatedAt       time.Time             `json:"created_at"`
}

type GetVacancyResponse struct {
	ID              uuid.UUID `json:"id"`
	Title           string    `json:"title"`
	KeyRequirements []string  `json:"key_requirements"`
	CreatedAt       time.Time `json:"created_at"`
}
