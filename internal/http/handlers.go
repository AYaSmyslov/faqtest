package httpapi

import (
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"strings"

	"github.com/AYaSmyslov/faqapi/internal/service"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

type createQuestionRequest struct {
	Text string `json:"text" example:"New question"`
}

type createAnswerRequest struct {
	UserID string `json:"user_id" example:"USER_NAME"`
	Text   string `json:"text" example:"New Answer"`
}

type methodMux map[string]http.HandlerFunc

func (m methodMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if handler := m[r.Method]; handler != nil {
		handler(w, r)
		return
	}

	allow := make([]string, 0, len(m)+1)
	for method := range m {
		allow = append(allow, method)
	}

	allow = append(allow, http.MethodOptions)
	sort.Strings(allow)
	w.Header().Set("Allow", strings.Join(allow, ", "))

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

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
	s.mux.Handle("/swagger/", httpSwagger.WrapHandler)

	s.mux.Handle("/questions", methodMux{
		http.MethodGet:  s.listQuestions,
		http.MethodPost: s.createQuestion,
	})

	s.mux.Handle("/questions/{questionID}", methodMux{
		http.MethodGet:    s.getQuestion,
		http.MethodDelete: s.deleteQuestion,
	})

	s.mux.Handle("/questions/{questionID}/answers", methodMux{
		http.MethodPost: s.createAnswer,
	})

	s.mux.Handle("/answers/{answerID}", methodMux{
		http.MethodGet:    s.getAnswer,
		http.MethodDelete: s.deleteAnswer,
	})
}

// createQuestion godoc
// @Summary Create question
// @Description Create and returns created question
// @Tags questions
// @Accept json
// @Produce json
// @Param request body createQuestionRequest true "Question payload"
// @Success 201 {object} models.Question
// @Failure 400 {object} map[string]string
// @Router /questions [post]
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

// listQuestions godoc
// @Summary List questions
// @Description Returns questions
// @Tags questions
// @Produce json
// @Success 200 {array} models.Question
// @Failure 500 {object} map[string]string
// @Router /questions [get]
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

// getQuestion godoc
// @Summary Get question by ID
// @Tags questions
// @Produce json
// @Param questionID path int true "Question ID"
// @Success 200 {object} models.Question
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /questions/{questionID} [get]
func (s *Server) getQuestion(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	questionID, err := pathInt64(r, "questionID")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid question id")
		return
	}

	question, err := s.svc.GetQuestionWithAnswers(ctx, uint(questionID))

	if err != nil {
		status := statusFromError(err)
		http.Error(w, err.Error(), status)
		return
	}

	writeJSON(w, http.StatusOK, question)
}

// deleteQuestion godoc
// @Summary Delete question
// @Description Delete question by ID
// @Tags questions
// @Param questionID path int true "Question ID"
// @Success 204
// @Failure 400 {object} map[string]string
// @Router /questions/{questionID} [delete]
func (s *Server) deleteQuestion(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	questionID, err := pathInt64(r, "questionID")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid question id")
		return
	}

	if err := s.svc.DeleteQuestion(ctx, uint(questionID)); err != nil {
		status := statusFromError(err)
		http.Error(w, err.Error(), status)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// createAnswer godoc
// @Summary Create answer
// @Description Create and returns created answer
// @Tags answers
// @Accept json
// @Produce json
// @Param questionID path int true "Question ID"
// @Param request body createAnswerRequest true "Answer payload"
// @Success 201 {object} models.Answer
// @Failure 400 {object} map[string]string
// @Router /questions/{questionID}/answers [post]
func (s *Server) createAnswer(w http.ResponseWriter, r *http.Request) {
	var req createAnswerRequest

	questionID, err := pathInt64(r, "questionID")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid question id")
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	answer, err := s.svc.CreateAnswer(ctx, uint(questionID), req.UserID, req.Text)
	if err != nil {
		status := statusFromError(err)
		http.Error(w, err.Error(), status)
		return
	}

	writeJSON(w, http.StatusCreated, answer)
}

// getAnswer godoc
// @Summary  Get answer by ID
// @Tags answers
// @Produce json
// @Param answerID path int true "Answer ID"
// @Success 200 {object} models.Answer
// @Failure 400 {object} map[string]string
// @Router /answers/{answerID} [get]
func (s *Server) getAnswer(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	answerID, err := pathInt64(r, "answerID")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid answer id")
		return
	}

	answer, err := s.svc.GetAnswer(ctx, uint(answerID))

	if err != nil {
		status := statusFromError(err)
		http.Error(w, err.Error(), status)
		return
	}

	writeJSON(w, http.StatusOK, answer)
}

// deleteAnswer godoc
// @Summary Delete answer
// @Description Delete answer by ID
// @Tags answers
// @Param answerID path int true "Answer ID"
// @Success 204
// @Failure 400 {object} map[string]string
// @Router /answers/{answerID} [delete]
func (s *Server) deleteAnswer(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	answerID, err := pathInt64(r, "answerID")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid answer id")
		return
	}

	if err := s.svc.DeleteAnswer(ctx, uint(answerID)); err != nil {
		status := statusFromError(err)
		http.Error(w, err.Error(), status)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
