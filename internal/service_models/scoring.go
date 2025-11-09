package service_models

import "hr-helper/internal/entity"

type ResumeScreeningResult struct {
	Score    int    `json:"score"`
	Feedback string `json:"feedback"`
}

type ResumeScreeningResultWithStatus struct {
	ResumeScreeningResult
	Status string
}

type AnswerScoringResult struct {
	Score int `json:"score"`
}

type ScoredAnswer struct {
	CandidateID int64
	QuestionID  int64
	Content     string
	TimeTaken   int
	Score       int
}

type InterviewResult struct {
	Status entity.CandidateVacancyStatus
	Score  int
}
