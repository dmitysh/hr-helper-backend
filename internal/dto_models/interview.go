package dto_models

import (
	"time"

	"github.com/google/uuid"
)

type CreateQuestionRequest struct {
	VacancyID uuid.UUID `json:"vacancy_id"`
	Content   string    `json:"content"`
	Reference string    `json:"reference"`
	TimeLimit int       `json:"time_limit"`
	Position  int       `json:"position"`
}

type CreateAnswerRequest struct {
	CandidateID int64  `json:"candidate_id"`
	QuestionID  int64  `json:"question_id"`
	Content     string `json:"content"`
	TimeTaken   int    `json:"time_taken"`
}
type GetQuestionsResponse struct {
	Questions []GetQuestionResponse `json:"questions"`
}

type GetQuestionResponse struct {
	ID        int64     `json:"id"`
	VacancyID int64     `json:"vacancy_id"`
	Content   string    `json:"content"`
	Reference string    `json:"reference"`
	TimeLimit int       `json:"time_limit"`
	Position  int       `json:"position"`
	CreatedAt time.Time `json:"created_at"`
}
