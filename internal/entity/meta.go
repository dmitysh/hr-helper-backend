package entity

import (
	"time"

	"github.com/google/uuid"
)

type CandidateVacancyStatus string

const (
	// CandidateVacancyStatusScreeningInProgress = "screening_in_progress"

	//CandidateVacancyStatusNew             = "new"
	CandidateVacancyStatusScreeningOk     = "screening_ok"
	CandidateVacancyStatusScreeningFailed = "screening_failed"
	CandidateVacancyStatusInterviewOk     = "interview_ok"
	CandidateVacancyStatusInterviewFailed = "interview_failed"
)

type Meta struct {
	CandidateID    int64                  `db:"candidate_id"`
	VacancyID      uuid.UUID              `db:"vacancy_id"`
	InterviewScore *int                   `db:"interview_score"`
	Status         CandidateVacancyStatus `db:"status"`
	UpdatedAt      time.Time              `db:"updated_at"`
	IsArchived     bool                   `db:"is_archived"`
}

type CandidateVacancyInfo struct {
	Candidate       Candidate
	Vacancy         Vacancy
	Meta            Meta
	ResumeScreening ResumeScreening
	Questions       []Question
	ResumeLink      string
}

type CandidateQuestionAnswer struct {
	Question Question
	Answer   Answer
}
