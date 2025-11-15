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
	"github.com/go-chi/cors"
	"github.com/google/uuid"

	"hr-helper/internal/dto_models"
	"hr-helper/internal/entity"
	"hr-helper/internal/inerrors"
	"hr-helper/internal/pkg/houston/loggy"
	"hr-helper/internal/service/candidate"
	"hr-helper/internal/service/vacancy"
)

type Server struct {
	httpServer       *http.Server
	candidateService *candidate.Service
	vacancyService   *vacancy.Service
}

func NewServer(addr string, candidateService *candidate.Service, vacancyService *vacancy.Service) *Server {
	s := &Server{
		httpServer: &http.Server{
			Addr: addr,
		},
		candidateService: candidateService,
		vacancyService:   vacancyService,
	}
	s.initHandlers()

	return s
}

func (s *Server) initHandlers() {
	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"}, // разрешённые домены
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // кэширование preflight запроса в секундах
	}))
	r.Use(middleware.Logger)

	r.Post("/api/v1/candidate", s.createCandidate)
	r.Post("/api/v1/screening/process", s.processResume)
	r.Post("/api/v1/answer", s.createAnswer)
	r.Post("/api/v1/interview/process", s.processInterview)
	r.Post("/api/v1/vacancy", s.createVacancy)
	r.Post("/api/v1/vacancy/archive", s.archiveVacancy)

	r.Delete("/api/v1/vacancy/{vacancy-id}", s.deleteVacancy)
	r.Delete("/api/_private/v1/candidate/{candidate-id}", s.deleteCandidate)

	r.Get("/api/v1/screening/result/{candidate-id}/{vacancy-id}", s.getScreeningResult)
	r.Get("/api/v1/meta/{candidate-id}/{vacancy-id}", s.getMeta)
	r.Get("/api/v1/candidates/by-tg-id/{telegram-id}", s.getCandidateByTelegramID)
	r.Get("/api/v1/questions/{vacancy-id}", s.getQuestionsByVacancyID)
	r.Get("/api/v1/candidate-vacancy-infos", s.getCandidateVacancyInfos)
	r.Get("/api/v1/vacancies", s.getVacancies)
	r.Get("/api/v1/vacancy/{vacancy-id}", s.getVacancyWithQuestionsByID)
	r.Get("/api/v1/candidate-vacancy-info/{candidate-id}/{vacancy-id}", s.getCandidateVacancyInfo)
	r.Get("/api/v1/candidate/answers/{candidate-id}/{vacancy-id}", s.getCandidateAnswers)

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

func (s *Server) deleteCandidate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	candidateID, err := strconv.ParseInt(chi.URLParam(r, "candidate-id"), 10, 64)
	if err != nil {
		httpErrorf(w, http.StatusBadRequest, "invalid candidate id: %v", err)
		return
	}

	err = s.candidateService.Delete(ctx, candidateID)
	if err != nil {
		httpErrorf(w, http.StatusInternalServerError, "can't handle delete: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) createVacancy(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var in dto_models.CreateVacancyRequest
	err := json.NewDecoder(r.Body).Decode(&in)
	if err != nil {
		httpErrorf(w, http.StatusBadRequest, "invalid JSON: %v", err.Error())
		return
	}

	id, err := s.vacancyService.CreateVacancy(ctx, in)
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

func (s *Server) archiveVacancy(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var in dto_models.ArchiveVacancyRequest
	err := json.NewDecoder(r.Body).Decode(&in)
	if err != nil {
		httpErrorf(w, http.StatusBadRequest, "invalid JSON: %v", err.Error())
		return
	}

	err = s.vacancyService.ArchiveVacancy(ctx, in.CandidateID, in.VacancyID)
	if err != nil {
		httpErrorf(w, http.StatusInternalServerError, "can't handle archive: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) createAnswer(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var in dto_models.CreateAnswerRequest
	err := json.NewDecoder(r.Body).Decode(&in)
	if err != nil {
		httpErrorf(w, http.StatusBadRequest, "invalid JSON: %v", err.Error())
		return
	}

	id, err := s.vacancyService.CreateAnswer(ctx, in)
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

func (s *Server) getMeta(w http.ResponseWriter, r *http.Request) {
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

	resumeScreeningResult, err := s.candidateService.GetMeta(ctx, candidateID, vacancyID)
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
	_ = json.NewEncoder(w).Encode(entityMetaToDTO(resumeScreeningResult))
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

	questions, err := s.vacancyService.GetQuestionsByVacancyID(ctx, vacancyID)
	if err != nil {
		httpErrorf(w, http.StatusInternalServerError, "can't handle get: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	var resp []dto_models.GetQuestionResponse
	for _, question := range questions {
		resp = append(resp, entityQuestionToDTO(question))
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

func (s *Server) processInterview(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var in dto_models.ProcessInterviewRequest
	err := json.NewDecoder(r.Body).Decode(&in)
	if err != nil {
		httpErrorf(w, http.StatusBadRequest, "invalid JSON: %v", err.Error())
		return
	}

	err = s.vacancyService.ScoreCandidateInterview(ctx, in)
	if errors.Is(err, inerrors.ErrNotFound) {
		httpErrorf(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		httpErrorf(w, http.StatusInternalServerError, "can't handle scoring: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) getCandidateVacancyInfos(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	candidateVacancyInfos, err := s.candidateService.GetCandidateVacancyInfos(ctx)
	if err != nil {
		httpErrorf(w, http.StatusInternalServerError, "can't handle get: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(entityCandidateVacancyInfosToDTO(candidateVacancyInfos))
}

func (s *Server) getVacancyWithQuestionsByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vacancyIDStr := chi.URLParam(r, "vacancy-id")
	vacancyID, err := uuid.Parse(vacancyIDStr)
	if err != nil {
		httpErrorf(w, http.StatusBadRequest, "invalid vacancy id")
		return
	}

	vacancy, err := s.vacancyService.GetVacancyWithQuestionsByID(ctx, vacancyID)
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
	_ = json.NewEncoder(w).Encode(entityVacancyWithAnswersToDTO(vacancy))
}

func (s *Server) deleteVacancy(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vacancyIDStr := chi.URLParam(r, "vacancy-id")
	vacancyID, err := uuid.Parse(vacancyIDStr)
	if err != nil {
		httpErrorf(w, http.StatusBadRequest, "invalid vacancy id")
		return
	}

	err = s.vacancyService.DeleteVacancy(ctx, vacancyID)
	if err != nil {
		httpErrorf(w, http.StatusInternalServerError, "can't handle delete: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) getCandidateVacancyInfo(w http.ResponseWriter, r *http.Request) {
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

	candidate, err := s.candidateService.GetCandidateVacancyInfo(ctx, candidateID, vacancyID)
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
	_ = json.NewEncoder(w).Encode(entityCandidateVacancyInfoToDTO(candidate))
}

func (s *Server) getCandidateAnswers(w http.ResponseWriter, r *http.Request) {
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

	answers, err := s.candidateService.GetCandidateAnswers(ctx, candidateID, vacancyID)
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
	_ = json.NewEncoder(w).Encode(entityCandidateQuestionAnswersToDTO(answers))
}

func (s *Server) getVacancies(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vacancies, err := s.vacancyService.GetVacanciesWithQuestions(ctx)
	if err != nil {
		httpErrorf(w, http.StatusInternalServerError, "can't handle get: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(entityVacanciesWithAnswersToDTO(vacancies))
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
		ID:               e.ID,
		TelegramID:       e.TelegramID,
		TelegramUsername: e.TelegramUsername,
		FullName:         e.FullName,
		Phone:            e.Phone,
		City:             e.City,
		CreatedAt:        e.CreatedAt,
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

func entityMetaToDTO(e entity.Meta) dto_models.GetMetaResponse {
	return dto_models.GetMetaResponse{
		CandidateID:    e.CandidateID,
		VacancyID:      e.VacancyID,
		InterviewScore: e.InterviewScore,
		Status:         string(e.Status),
		IsArchived:     e.IsArchived,
		UpdatedAt:      e.UpdatedAt,
	}
}

func entityVacanciesWithAnswersToDTO(es []entity.VacancyWithQuestion) []dto_models.GetVacancyWithQuestionsResponse {
	res := make([]dto_models.GetVacancyWithQuestionsResponse, 0, len(es))

	for _, e := range es {
		res = append(res, entityVacancyWithAnswersToDTO(e))
	}

	return res
}
func entityVacancyWithAnswersToDTO(e entity.VacancyWithQuestion) dto_models.GetVacancyWithQuestionsResponse {
	v := dto_models.GetVacancyWithQuestionsResponse{
		ID:              e.ID,
		Title:           e.Title,
		KeyRequirements: e.KeyRequirements,
		Questions:       make([]dto_models.GetQuestionResponse, 0, len(e.Questions)),
		CreatedAt:       e.CreatedAt,
	}
	for _, q := range e.Questions {
		v.Questions = append(v.Questions, dto_models.GetQuestionResponse{
			ID:        q.ID,
			VacancyID: q.VacancyID,
			Content:   q.Content,
			Reference: q.Reference,
			TimeLimit: q.TimeLimit,
			Position:  q.Position,
			CreatedAt: q.CreatedAt,
		})
	}
	return v
}

func entityCandidateVacancyInfosToDTO(es []entity.CandidateVacancyInfo) []dto_models.GetCandidateVacancyInfoResponse {
	res := make([]dto_models.GetCandidateVacancyInfoResponse, 0, len(es))

	for _, e := range es {
		res = append(res, entityCandidateVacancyInfoToDTO(e))
	}

	return res
}

func entityCandidateVacancyInfoToDTO(e entity.CandidateVacancyInfo) dto_models.GetCandidateVacancyInfoResponse {
	return dto_models.GetCandidateVacancyInfoResponse{
		Candidate: dto_models.GetCandidateResponse{
			ID:               e.Candidate.ID,
			TelegramID:       e.Candidate.TelegramID,
			TelegramUsername: e.Candidate.TelegramUsername,
			FullName:         e.Candidate.FullName,
			Phone:            e.Candidate.Phone,
			City:             e.Candidate.City,
			CreatedAt:        e.Candidate.CreatedAt,
		},
		Vacancy: dto_models.GetVacancyResponse{
			ID:              e.Vacancy.ID,
			Title:           e.Vacancy.Title,
			KeyRequirements: e.Vacancy.KeyRequirements,
			CreatedAt:       e.Vacancy.CreatedAt,
		},
		Meta: dto_models.GetMetaResponse{
			CandidateID:    e.Candidate.ID,
			VacancyID:      e.Vacancy.ID,
			InterviewScore: e.Meta.InterviewScore,
			Status:         string(e.Meta.Status),
			IsArchived:     e.Meta.IsArchived,
			UpdatedAt:      e.Meta.UpdatedAt,
		},
		ResumeScreening: entityResumeScreeningToDTO(e.ResumeScreening),
		ResumeLink:      e.ResumeLink,
	}
}

func entityCandidateQuestionAnswersToDTO(es []entity.CandidateQuestionAnswer) []dto_models.GetCandidateQuestionAnswerResponse {
	res := make([]dto_models.GetCandidateQuestionAnswerResponse, 0, len(es))

	for _, e := range es {
		res = append(res, dto_models.GetCandidateQuestionAnswerResponse{
			Question: dto_models.GetQuestionResponse{
				ID:        e.Question.ID,
				VacancyID: e.Question.VacancyID,
				Content:   e.Question.Content,
				Reference: e.Question.Reference,
				TimeLimit: e.Question.TimeLimit,
				Position:  e.Question.Position,
				CreatedAt: e.Question.CreatedAt,
			},
			Answer: dto_models.GetAnswerResponse{
				ID:          e.Answer.ID,
				CandidateID: e.Answer.CandidateID,
				QuestionID:  e.Answer.QuestionID,
				Content:     e.Answer.Content,
				Score:       e.Answer.Score,
				TimeTaken:   e.Answer.TimeTaken,
				CreatedAt:   e.Answer.CreatedAt,
			},
		})
	}

	return res
}
