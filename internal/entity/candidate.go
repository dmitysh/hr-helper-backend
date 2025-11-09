package entity

import (
	"time"

	"github.com/google/uuid"
)

type Candidate struct {
	ID         int64     `db:"id"`
	TelegramID int64     `db:"telegram_id"`
	FullName   string    `db:"full_name"`
	Phone      string    `db:"phone"`
	City       string    `db:"city"`
	CreatedAt  time.Time `db:"created_at"`
}

type CandidateDetail struct {
	Candidate               Candidate
	CandidateVacancyDetails []CandidateVacancyDetail
}

type CandidateVacancyDetail struct {
	ID              uuid.UUID
	Title           string
	ResumeScreening ResumeScreening
	Questions       []Question
	Answers         []Answer
	CreatedAt       time.Time
}

type ResumeScreening struct {
	ID          int64
	CandidateID int64
	VacancyID   uuid.UUID
	Score       int
	Feedback    string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
