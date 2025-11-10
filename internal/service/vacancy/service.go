package vacancy

import (
	"context"
	"fmt"
	"math"

	"github.com/google/uuid"

	"hr-helper/internal/dto_models"
	"hr-helper/internal/entity"
	"hr-helper/internal/service_models"
)

const (
	minInterviewScore = 75
)

type Storage interface {
	CreateVacancy(ctx context.Context, vacancy dto_models.CreateVacancyRequest) (uuid.UUID, error)
	CreateAnswer(ctx context.Context, answer service_models.ScoredAnswer) (int64, error)
	GetQuestionByID(ctx context.Context, id int64) (entity.Question, error)
	GetQuestionsByVacancyID(ctx context.Context, vacancyID uuid.UUID) ([]entity.Question, error)
	GetAnswers(ctx context.Context, candidateID int64, vacancyID uuid.UUID) ([]entity.Answer, error)
	GetVacanciesWithQuestions(ctx context.Context) ([]entity.VacancyWithQuestion, error)
	GetVacancyWithQuestions(ctx context.Context, vacancyID uuid.UUID) (entity.VacancyWithQuestion, error)
	UpdateInterviewResult(ctx context.Context, candidateID int64, vacancyID uuid.UUID, interviewResult service_models.InterviewResult) error
	DeleteVacancy(ctx context.Context, vacancyID uuid.UUID) error
}

type LLMClient interface {
	ScoreAnswer(ctx context.Context, answer string, reference string) (service_models.AnswerScoringResult, error)
}

type Service struct {
	store     Storage
	llmClient LLMClient
}

func NewService(store Storage, llmClient LLMClient) *Service {
	return &Service{
		store:     store,
		llmClient: llmClient,
	}
}

func (s *Service) CreateVacancy(ctx context.Context, vacancy dto_models.CreateVacancyRequest) (uuid.UUID, error) {
	return s.store.CreateVacancy(ctx, vacancy)
}

func (s *Service) CreateAnswer(ctx context.Context, req dto_models.CreateAnswerRequest) (int64, error) {
	question, err := s.store.GetQuestionByID(ctx, req.QuestionID)
	if err != nil {
		return 0, fmt.Errorf("can't get question: %w", err)
	}

	scoringResult, err := s.llmClient.ScoreAnswer(ctx, req.Content, question.Reference)
	if err != nil {
		return 0, fmt.Errorf("can't score answer via llm: %w", err)
	}

	id, err := s.store.CreateAnswer(ctx, service_models.ScoredAnswer{
		CandidateID: req.CandidateID,
		QuestionID:  req.QuestionID,
		Content:     req.Content,
		TimeTaken:   req.TimeTaken,
		Score:       scoringResult.Score,
	})
	if err != nil {
		return 0, fmt.Errorf("can't create answer in db: %w", err)
	}

	return id, nil
}

func (s *Service) GetQuestionsByVacancyID(ctx context.Context, vacancyID uuid.UUID) ([]entity.Question, error) {
	questions, err := s.store.GetQuestionsByVacancyID(ctx, vacancyID)
	if err != nil {
		return nil, fmt.Errorf("can't get questions: %w", err)
	}

	return questions, nil
}

func (s *Service) DeleteVacancy(ctx context.Context, vacancyID uuid.UUID) error {
	err := s.store.DeleteVacancy(ctx, vacancyID)
	if err != nil {
		return fmt.Errorf("can't delete vacancy: %w", err)
	}

	return nil
}

func (s *Service) GetVacanciesWithQuestions(ctx context.Context) ([]entity.VacancyWithQuestion, error) {
	vacancies, err := s.store.GetVacanciesWithQuestions(ctx)
	if err != nil {
		return nil, fmt.Errorf("can't get vacancies: %w", err)
	}

	return vacancies, nil
}

func (s *Service) GetVacancyWithQuestionsByID(ctx context.Context, vacancyID uuid.UUID) (entity.VacancyWithQuestion, error) {
	vacancy, err := s.store.GetVacancyWithQuestions(ctx, vacancyID)
	if err != nil {
		return entity.VacancyWithQuestion{}, fmt.Errorf("can't get vacancy: %w", err)
	}

	return vacancy, nil
}

func (s *Service) ScoreCandidateInterview(ctx context.Context, req dto_models.ProcessInterviewRequest) error {
	answers, err := s.store.GetAnswers(ctx, req.CandidateID, req.VacancyID)
	if err != nil {
		return fmt.Errorf("can't get answers: %w", err)
	}

	res := s.checkInterviewScore(answers)

	err = s.store.UpdateInterviewResult(ctx, req.CandidateID, req.VacancyID, res)
	if err != nil {
		return fmt.Errorf("can't update scoring results: %w", err)
	}

	return nil
}

func (s *Service) checkInterviewScore(answers []entity.Answer) service_models.InterviewResult {
	scoreSum := 0.0
	for _, answer := range answers {
		scoreSum += float64(answer.Score)
	}

	avgScore := int(math.Round(scoreSum / float64(len(answers))))
	res := service_models.InterviewResult{
		Score: avgScore,
	}
	if avgScore >= minInterviewScore {
		res.Status = entity.CandidateVacancyStatusInterviewOk
	} else {
		res.Status = entity.CandidateVacancyStatusInterviewFailed
	}

	return res
}
