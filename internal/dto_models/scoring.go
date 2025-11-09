package dto_models

import "github.com/google/uuid"

type ProcessResumeRequest struct {
	CandidateID int64     `json:"candidate_id"`
	VacancyID   uuid.UUID `json:"vacancy_id"`
}

type ProcessInterviewRequest struct {
	CandidateID int64     `json:"candidate_id"`
	VacancyID   uuid.UUID `json:"vacancy_id"`
}
