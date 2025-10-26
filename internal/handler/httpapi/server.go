package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"

	"hr-helper/internal/dto_models"
	"hr-helper/internal/entity"
	"hr-helper/internal/inerrors"
	"hr-helper/internal/pkg/houston/loggy"
	"hr-helper/internal/service/candidate"
	"hr-helper/internal/service/interview"
	"hr-helper/internal/service/vacancy"
)

type Server struct {
	httpServer       *http.Server
	candidateService *candidate.Service
	interviewService *interview.Service
	vacancyService   *vacancy.Service
}

func NewServer(addr string, candidateService *candidate.Service, interviewService *interview.Service, vacancyService *vacancy.Service) *Server {
	s := &Server{
		httpServer: &http.Server{
			Addr: addr,
		},
		candidateService: candidateService,
		interviewService: interviewService,
		vacancyService:   vacancyService,
	}
	s.initHandlers()

	return s
}

func (s *Server) initHandlers() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Post("/api/v1/candidate", s.createCandidate)
	r.Post("/api/v1/screening/process", s.processResume)
	r.Get("/api/v1/screening/result/{candidate-id}/{vacancy-id}", s.getScreeningResult)
	r.Get("/api/v1/candidates/by-tg-id/{telegram-id}", s.getCandidateByTelegramID)
	r.Post("/api/v1/question", s.createQuestion)
	r.Get("/api/v1/questions/{vacancy-id}", s.getQuestionsByVacancyID)
	r.Post("/api/v1/answer", s.createAnswer)
	r.Post("/api/v1/vacancy", s.createVacancy)

	s.httpServer.Handler = r
}

func (s *Server) createCandidate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var in dto_models.CreateCandidateRequest
	err := json.NewDecoder(r.Body).Decode(&in)
	if err != nil {
		httpErrorf(w, http.StatusBadRequest, "invalid JSON: %v", err.Error())
		return
	}

	id, err := s.candidateService.Create(ctx, in)
	if err != nil {
		httpErrorf(w, http.StatusInternalServerError, "can't handle creation: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"id": id,
	})
}

func (s *Server) createQuestion(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var in dto_models.CreateQuestionRequest
	err := json.NewDecoder(r.Body).Decode(&in)
	if err != nil {
		httpErrorf(w, http.StatusBadRequest, "invalid JSON: %v", err.Error())
		return
	}

	id, err := s.interviewService.CreateQuestion(ctx, in)
	if err != nil {
		httpErrorf(w, http.StatusInternalServerError, "can't handle creation: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"id": id,
	})
}

func (s *Server) createVacancy(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var in dto_models.CreateVacancyRequest
	err := json.NewDecoder(r.Body).Decode(&in)
	if err != nil {
		httpErrorf(w, http.StatusBadRequest, "invalid JSON: %v", err.Error())
		return
	}

	id, err := s.vacancyService.Create(ctx, in)
	if err != nil {
		httpErrorf(w, http.StatusInternalServerError, "can't handle creation: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"id": id,
	})
}

func (s *Server) createAnswer(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var in dto_models.CreateAnswerRequest
	err := json.NewDecoder(r.Body).Decode(&in)
	if err != nil {
		httpErrorf(w, http.StatusBadRequest, "invalid JSON: %v", err.Error())
		return
	}

	id, err := s.interviewService.CreateAnswer(ctx, in)
	if err != nil {
		httpErrorf(w, http.StatusInternalServerError, "can't handle creation: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"id": id,
	})
}

func (s *Server) getCandidateByTelegramID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	telegramID, err := strconv.ParseInt(chi.URLParam(r, "telegram-id"), 10, 64)
	if err != nil {
		httpErrorf(w, http.StatusBadRequest, "invalid telegram id: %v", err)
		return
	}

	candidate, err := s.candidateService.GetByTelegramID(ctx, telegramID)
	if errors.Is(err, inerrors.ErrNotFound) {
		httpErrorf(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		httpErrorf(w, http.StatusInternalServerError, "can't handle get: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(entityCandidateToDTO(candidate))
}

func (s *Server) getScreeningResult(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	candidateID, err := strconv.ParseInt(chi.URLParam(r, "candidate-id"), 10, 64)
	if err != nil {
		httpErrorf(w, http.StatusBadRequest, "invalid candidate id: %v", err)
		return
	}

	vacancyIDStr := chi.URLParam(r, "vacancy-id")
	vacancyID, err := uuid.Parse(vacancyIDStr)
	if err != nil {
		httpErrorf(w, http.StatusBadRequest, "invalid vacancy id")
		return
	}

	resumeScreeningResult, err := s.candidateService.GetResumeScreening(ctx, candidateID, vacancyID)
	if errors.Is(err, inerrors.ErrNotFound) {
		httpErrorf(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		httpErrorf(w, http.StatusInternalServerError, "can't handle get: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(entityResumeScreeningToDTO(resumeScreeningResult))
}

func (s *Server) getQuestionsByVacancyID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vacancyIDStr := chi.URLParam(r, "vacancy-id")
	vacancyID, err := uuid.Parse(vacancyIDStr)
	if err != nil {
		httpErrorf(w, http.StatusBadRequest, "invalid vacancy id")
		return
	}

	questions, err := s.interviewService.GetQuestionsByVacancyID(ctx, vacancyID)
	if err != nil {
		httpErrorf(w, http.StatusInternalServerError, "can't handle get: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	var resp dto_models.GetQuestionsResponse
	for _, question := range questions {
		resp.Questions = append(resp.Questions, entityQuestionToDTO(question))
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *Server) processResume(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var in dto_models.ProcessResumeRequest
	err := json.NewDecoder(r.Body).Decode(&in)
	if err != nil {
		httpErrorf(w, http.StatusBadRequest, "invalid JSON: %v", err.Error())
		return
	}

	err = s.candidateService.ScoreCandidateResume(ctx, in)
	if errors.Is(err, inerrors.ErrNotFound) {
		httpErrorf(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		httpErrorf(w, http.StatusInternalServerError, "can't handle screening: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func httpError(w http.ResponseWriter, code int, msg string) {
	loggy.Errorf(msg)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": msg,
	})
}

func httpErrorf(w http.ResponseWriter, code int, format string, args ...any) {
	loggy.Errorf(format, args...)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": fmt.Sprintf(format, args...),
	})
}

func (s *Server) Start() error {
	loggy.Infof("starting http server on %s", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func entityCandidateToDTO(e entity.Candidate) dto_models.GetCandidateResponse {
	return dto_models.GetCandidateResponse{
		ID:         e.ID,
		TelegramID: e.TelegramID,
		FullName:   e.FullName,
		Phone:      e.Phone,
		City:       e.City,
		CreatedAt:  e.CreatedAt,
	}
}

func entityQuestionToDTO(e entity.Question) dto_models.GetQuestionResponse {
	return dto_models.GetQuestionResponse{
		ID:        e.ID,
		VacancyID: e.VacancyID,
		Content:   e.Content,
		Reference: e.Reference,
		TimeLimit: e.TimeLimit,
		Position:  e.Position,
		CreatedAt: e.CreatedAt,
	}
}

func entityResumeScreeningToDTO(e entity.ResumeScreening) dto_models.GetResumeScreeningResponse {
	return dto_models.GetResumeScreeningResponse{
		ID:          e.ID,
		CandidateID: e.CandidateID,
		VacancyID:   e.VacancyID,
		Score:       e.Score,
		Feedback:    e.Feedback,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}
}
