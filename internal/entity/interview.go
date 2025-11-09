package entity

import (
	"time"

	"github.com/google/uuid"
)

type Question struct {
	ID        int64     `db:"id" json:"id"`
	VacancyID uuid.UUID `db:"vacancy_id" json:"vacancy_id"`
	Content   string    `db:"content" json:"content"`
	Reference string    `db:"reference" json:"reference"`
	TimeLimit int       `db:"time_limit" json:"time_limit"`
	Position  int       `db:"position" json:"position"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type Answer struct {
	ID          int64     `db:"id"`
	CandidateID int64     `db:"candidate_id"`
	QuestionID  int64     `db:"question_id"`
	Content     string    `db:"content"`
	Score       int       `db:"score"`
	TimeTaken   int64     `db:"time_taken"`
	CreatedAt   time.Time `db:"created_at"`
}
