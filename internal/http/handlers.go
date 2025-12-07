package httpapi

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/AYaSmyslov/faqapi/internal/service"
)

type Server struct {
	svc *service.FAQService
	mux *http.ServeMux
}

func NewServer(svc *service.FAQService) *Server {
	server := &Server{
		svc: svc,
		mux: http.NewServeMux(),
	}

	server.routes()

	return server
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) routes() {
	s.mux.HandleFunc("/questions/", s.handleQuestions)
	s.mux.HandleFunc("/answers/", s.handleAnswers)
}

func (s *Server) handleQuestions(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/questions")

	if path == "" || path == "/" {
		switch r.Method {
		case http.MethodGet:
			s.listQuestions(w, r)
		case http.MethodPost:
			s.createQuestion(w, r)
		default:
			http.Error(w, "metod not allowed", http.StatusMethodNotAllowed)
		}

		return
	}

	parts := strings.Split(strings.Trim(path, "/"), "/")

	if len(parts) == 1 {
		// /questions/{id}
		id, err := strconv.ParseUint(parts[0], 10, 64)

		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodGet:
			s.getQuestion(w, r, uint(id))
		case http.MethodDelete:
			s.deleteQuestion(w, r, uint(id))
		default:
			http.Error(w, "metod not allowed", http.StatusMethodNotAllowed)

		}
		return
	}

	if len(parts) == 2 && parts[1] == "answers" {
		// /questions/{id}/answers
		id, err := strconv.ParseUint(parts[0], 10, 64)
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}

		if r.Method == http.MethodPost {
			s.createAnswer(w, r, uint(id))
			return
		}

		http.Error(w, "metod not allowed", http.StatusMethodNotAllowed)
		return
	}

	http.NotFound(w, r)
}

func (s *Server) handleAnswers(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/answers/")

	if path == "" {
		http.NotFound(w, r)
		return
	}

	id, err := strconv.ParseUint(strings.Trim(path, "/"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.getAnswer(w, r, uint(id))
	case http.MethodDelete:
		s.deleteAnswer(w, r, uint(id))
	default:
		http.Error(w, "metod not allowed", http.StatusMethodNotAllowed)
	}
}

type createQuestionRequest struct {
	Text string `json:"text"`
}

func (s *Server) createQuestion(w http.ResponseWriter, r *http.Request) {
	var req createQuestionRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	question, err := s.svc.CreateQuestion(ctx, req.Text)

	if err != nil {
		status := statusFromError(err)
		http.Error(w, err.Error(), status)
		return
	}

	writeJSON(w, http.StatusCreated, question)
}

func (s *Server) listQuestions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	questions, err := s.svc.ListQuestions(ctx)

	if err != nil {
		log.Printf("list questions error: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, questions)
}

func (s *Server) getQuestion(w http.ResponseWriter, r *http.Request, id uint) {
	ctx := r.Context()
	question, err := s.svc.GetQuestionWithAnswers(ctx, id)

	if err != nil {
		status := statusFromError(err)
		http.Error(w, err.Error(), status)
		return
	}

	writeJSON(w, http.StatusOK, question)
}

func (s *Server) deleteQuestion(w http.ResponseWriter, r *http.Request, id uint) {
	ctx := r.Context()

	if err := s.svc.DeleteQuestion(ctx, id); err != nil {
		status := statusFromError(err)
		http.Error(w, err.Error(), status)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type createAnswerRequest struct {
	UserID string `json:"user_id"`
	Text   string `json:"text"`
}

func (s *Server) createAnswer(w http.ResponseWriter, r *http.Request, questionID uint) {
	var req createAnswerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	answer, err := s.svc.CreateAnswer(ctx, questionID, req.UserID, req.Text)
	if err != nil {
		status := statusFromError(err)
		http.Error(w, err.Error(), status)
		return
	}

	writeJSON(w, http.StatusCreated, answer)
}

func (s *Server) getAnswer(w http.ResponseWriter, r *http.Request, id uint) {
	ctx := r.Context()
	answer, err := s.svc.GetAnswer(ctx, id)

	if err != nil {
		status := statusFromError(err)
		http.Error(w, err.Error(), status)
		return
	}

	writeJSON(w, http.StatusOK, answer)
}

func (s *Server) deleteAnswer(w http.ResponseWriter, r *http.Request, id uint) {
	ctx := r.Context()

	if err := s.svc.DeleteAnswer(ctx, id); err != nil {
		status := statusFromError(err)
		http.Error(w, err.Error(), status)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func statusFromError(err error) int {
	switch err {
	case service.ErrBadRequest:
		return http.StatusBadRequest
	case service.ErrNotFound, service.ErrNoSuchQuestion:
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
		log.Printf("%s %s done in %s", r.Method, r.URL.Path, time.Since(start))
	})
}
