package entity

import (
	"time"

	"github.com/google/uuid"
)

type Vacancy struct {
	ID              uuid.UUID
	Title           string
	KeyRequirements []string
	CreatedAt       time.Time
}
