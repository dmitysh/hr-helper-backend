package entity

import (
	"time"

	"github.com/google/uuid"
)

type Vacancy struct {
	ID              uuid.UUID `db:"id"`
	Title           string    `db:"title"`
	KeyRequirements []string  `db:"key_requirements"`
	CreatedAt       time.Time `db:"created_at"`
}

type VacancyWithQuestion struct {
	ID              uuid.UUID
	Title           string
	KeyRequirements []string
	Questions       []Question
	CreatedAt       time.Time
}
