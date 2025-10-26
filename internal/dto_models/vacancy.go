package dto_models

import "github.com/google/uuid"

type CreateVacancyRequest struct {
	ID              uuid.UUID `json:"id"`
	Title           string    `json:"title"`
	KeyRequirements []string  `json:"key_requirements"`
}
