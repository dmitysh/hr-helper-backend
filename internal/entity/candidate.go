package entity

import (
	"time"

	"github.com/google/uuid"
)

type Candidate struct {
	ID               int64
	TelegramID       int64
	TelegramUsername string
	FullName         string
	Phone            string
	City             string
	CreatedAt        time.Time
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
