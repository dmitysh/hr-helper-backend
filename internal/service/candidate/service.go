package candidate

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"

	"hr-helper/internal/dto_models"
	"hr-helper/internal/entity"
	"hr-helper/internal/service_models"
)

const (
	minResumeScore = 75
)

type Storage interface {
	Create(ctx context.Context, candidate dto_models.CreateCandidateRequest) (int64, error)
	GetByTelegramID(ctx context.Context, telegramID int64) (entity.Candidate, error)
	UpdateScreeningResult(ctx context.Context, candidateID int64, vacancyID uuid.UUID, result service_models.ResumeScreeningResultWithStatus) error
	GetResumeScreening(ctx context.Context, candidateID int64, vacancyID uuid.UUID) (entity.ResumeScreening, error)
	GetMeta(ctx context.Context, candidateID int64, vacancyID uuid.UUID) (entity.Meta, error)
	GetCandidateVacancyInfos(ctx context.Context) ([]entity.CandidateVacancyInfo, error)
	GetCandidateVacancyInfo(ctx context.Context, candidateID int64, vacancyID uuid.UUID) (entity.CandidateVacancyInfo, error)
	GetCandidateAnswers(ctx context.Context, candidateID int64, vacancyID uuid.UUID) ([]entity.CandidateQuestionAnswer, error)
	Delete(ctx context.Context, candidateID int64) error
}

type VacancyStorage interface {
	GetByID(ctx context.Context, id uuid.UUID) (entity.Vacancy, error)
}

type LLMClient interface {
	ScoreResume(ctx context.Context, resumeBase64 string, vacancy entity.Vacancy) (service_models.ResumeScreeningResult, error)
}

type ResumeStorage interface {
	Download(ctx context.Context, candidateID int64, vacancyID uuid.UUID) ([]byte, error)
	GetPresignedURL(ctx context.Context, candidateID int64, vacancyID uuid.UUID) (string, error)
}

type Service struct {
	tikaURL string

	store         Storage
	llmClient     LLMClient
	vacancyStore  VacancyStorage
	resumeStorage ResumeStorage
}

func NewService(tikaURL string, store Storage, resumeStorage ResumeStorage, vacancyStorage VacancyStorage, llmClient LLMClient) *Service {
	return &Service{
		tikaURL:       tikaURL,
		store:         store,
		resumeStorage: resumeStorage,
		vacancyStore:  vacancyStorage,
		llmClient:     llmClient,
	}
}

func (s *Service) Create(ctx context.Context, candidate dto_models.CreateCandidateRequest) (int64, error) {
	return s.store.Create(ctx, candidate)
}

func (s *Service) Delete(ctx context.Context, candidateID int64) error {
	return s.store.Delete(ctx, candidateID)
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

	resumeText, err := s.extractTextFromPDF(resumeBytes)
	if err != nil {
		return fmt.Errorf("can't extract text from resume: %w", err)
	}

	scoringResult, err := s.llmClient.ScoreResume(ctx, resumeText, vacancy)
	if err != nil {
		return fmt.Errorf("can't score resume via llm: %w", err)
	}

	scoringResultWithStatus := service_models.ResumeScreeningResultWithStatus{
		ResumeScreeningResult: scoringResult,
		Status:                s.checkScreeningScore(scoringResult.Score),
	}

	err = s.store.UpdateScreeningResult(ctx, req.CandidateID, req.VacancyID, scoringResultWithStatus)
	if err != nil {
		return fmt.Errorf("can't update scoring results: %w", err)
	}

	return nil
}

func (s *Service) extractTextFromPDF(pdfData []byte) (string, error) {
	req, err := http.NewRequest("PUT", s.tikaURL+"/tika", bytes.NewReader(pdfData))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/pdf")
	req.Header.Set("Accept", "text/plain")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func (s *Service) checkScreeningScore(score int) string {
	if score >= minResumeScore {
		return entity.CandidateVacancyStatusScreeningOk
	}

	return entity.CandidateVacancyStatusScreeningFailed
}

func (s *Service) GetResumeScreening(ctx context.Context, candidateID int64, vacancyID uuid.UUID) (entity.ResumeScreening, error) {
	return s.store.GetResumeScreening(ctx, candidateID, vacancyID)
}

func (s *Service) GetMeta(ctx context.Context, candidateID int64, vacancyID uuid.UUID) (entity.Meta, error) {
	return s.store.GetMeta(ctx, candidateID, vacancyID)
}

func (s *Service) GetCandidateVacancyInfo(ctx context.Context, candidateID int64, vacancyID uuid.UUID) (entity.CandidateVacancyInfo, error) {
	info, err := s.store.GetCandidateVacancyInfo(ctx, candidateID, vacancyID)
	if err != nil {
		return entity.CandidateVacancyInfo{}, fmt.Errorf("can't get info: %w", err)
	}

	resumeLink, err := s.resumeStorage.GetPresignedURL(ctx, candidateID, vacancyID)
	if err != nil {
		return entity.CandidateVacancyInfo{}, fmt.Errorf("can't get resume link: %w", err)
	}
	info.ResumeLink = resumeLink

	return info, nil
}

func (s *Service) GetCandidateVacancyInfos(ctx context.Context) ([]entity.CandidateVacancyInfo, error) {
	return s.store.GetCandidateVacancyInfos(ctx)
}

func (s *Service) GetCandidateAnswers(ctx context.Context, candidateID int64, vacancyID uuid.UUID) ([]entity.CandidateQuestionAnswer, error) {
	return s.store.GetCandidateAnswers(ctx, candidateID, vacancyID)
}
