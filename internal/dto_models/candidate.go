package dto_models

import (
	"time"

	"github.com/google/uuid"
)

type CreateCandidateRequest struct {
	TelegramID int64  `json:"telegram_id"`
	FullName   string `json:"full_name"`
	Phone      string `json:"phone"`
	City       string `json:"city"`
}

type GetCandidateResponse struct {
	ID         int64     `json:"id"`
	TelegramID int64     `json:"telegram_id"`
	FullName   string    `json:"full_name"`
	Phone      string    `json:"phone"`
	City       string    `json:"city"`
	CreatedAt  time.Time `json:"created_at"`
}

type GetResumeScreeningResponse struct {
	ID          int64     `json:"id"`
	CandidateID int64     `json:"candidate_id"`
	VacancyID   uuid.UUID `json:"vacancy_id"`
	Score       int       `json:"score"`
	Feedback    string    `json:"feedback"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
