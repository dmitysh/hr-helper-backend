package service_models

type ResumeScreeningResult struct {
	Score    int    `json:"score"`
	Feedback string `json:"feedback"`
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
