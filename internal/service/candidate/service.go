package candidate

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/google/uuid"

	"hr-helper/internal/dto_models"
	"hr-helper/internal/entity"
	"hr-helper/internal/service_models"
)

type Storage interface {
	Create(ctx context.Context, candidate dto_models.CreateCandidateRequest) (int64, error)
	GetByTelegramID(ctx context.Context, telegramID int64) (entity.Candidate, error)
	UpdateScreeningResult(ctx context.Context, candidateID int64, vacancyID uuid.UUID, result service_models.ResumeScreeningResult) error
	GetResumeScreening(ctx context.Context, candidateID int64, vacancyID uuid.UUID) (entity.ResumeScreening, error)
}

type VacancyStorage interface {
	GetByID(ctx context.Context, id uuid.UUID) (entity.Vacancy, error)
}

type LLMClient interface {
	ScoreResume(ctx context.Context, resumeBase64 string, vacancy entity.Vacancy) (service_models.ResumeScreeningResult, error)
}

type ResumeStorage interface {
	Download(ctx context.Context, candidateID int64, vacancyID uuid.UUID) ([]byte, error)
}

type Service struct {
	store         Storage
	llmClient     LLMClient
	vacancyStore  VacancyStorage
	resumeStorage ResumeStorage
}

func NewService(store Storage, resumeStorage ResumeStorage, vacancyStorage VacancyStorage, llmClient LLMClient) *Service {
	return &Service{
		store:         store,
		resumeStorage: resumeStorage,
		vacancyStore:  vacancyStorage,
		llmClient:     llmClient,
	}
}

func (s *Service) Create(ctx context.Context, candidate dto_models.CreateCandidateRequest) (int64, error) {
	return s.store.Create(ctx, candidate)
}

func (s *Service) GetByTelegramID(ctx context.Context, telegramID int64) (entity.Candidate, error) {
	return s.store.GetByTelegramID(ctx, telegramID)
}

func (s *Service) ScoreCandidateResume(ctx context.Context, req dto_models.ProcessResumeRequest) error {
	vacancy, err := s.vacancyStore.GetByID(ctx, req.VacancyID)
	if err != nil {
		return fmt.Errorf("can't get vacancy: %w", err)
	}

	resumeBytes, err := s.resumeStorage.Download(ctx, req.CandidateID, req.VacancyID)
	if err != nil {
		return fmt.Errorf("can't download resume: %w", err)
	}

	resumeB64 := base64.StdEncoding.EncodeToString(resumeBytes)
	scoringResult, err := s.llmClient.ScoreResume(ctx, resumeB64, vacancy)
	if err != nil {
		return fmt.Errorf("can't score resume via llm: %w", err)
	}

	err = s.store.UpdateScreeningResult(ctx, req.CandidateID, req.VacancyID, scoringResult)
	if err != nil {
		return fmt.Errorf("can't update scoring results: %w", err)
	}

	return nil
}

func (s *Service) GetResumeScreening(ctx context.Context, candidateID int64, vacancyID uuid.UUID) (entity.ResumeScreening, error) {
	return s.store.GetResumeScreening(ctx, candidateID, vacancyID)
}
