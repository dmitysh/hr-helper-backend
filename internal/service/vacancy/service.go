package vacancy

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"hr-helper/internal/dto_models"
	"hr-helper/internal/entity"
	"hr-helper/internal/service_models"
)

type Storage interface {
	CreateVacancy(ctx context.Context, vacancy dto_models.CreateVacancyRequest) (uuid.UUID, error)
	CreateAnswer(ctx context.Context, answer service_models.ScoredAnswer) (int64, error)
	GetQuestionByID(ctx context.Context, id int64) (entity.Question, error)
	GetQuestionsByVacancyID(ctx context.Context, vacancyID uuid.UUID) ([]entity.Question, error)
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
