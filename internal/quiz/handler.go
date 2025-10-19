package quiz

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/saulo-duarte/chronos-lambda/internal/auth"
	"github.com/saulo-duarte/chronos-lambda/internal/config"
)

type Handler struct {
	service QuizService
}

func NewHandler(s QuizService) *Handler {
	return &Handler{service: s}
}

func (h *Handler) CreateQuiz(w http.ResponseWriter, r *http.Request) {
	log := config.WithContext(r.Context())

	claims, err := auth.GetUserClaimsFromContext(r.Context())
	if err != nil {
		log.Warn("Usuário não autenticado para criar quiz")
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var payload struct {
		Quiz      Quiz            `json:"quiz"`
		Questions []*QuizQuestion `json:"questions"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		log.WithError(err).Error("Corpo da requisição inválido para criar quiz")
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if len(payload.Questions) == 0 {
		log.Warn("Tentativa de criar quiz sem perguntas")
		http.Error(w, "quiz must contain at least one question", http.StatusBadRequest)
		return
	}

	payload.Quiz.UserID = uuid.MustParse(claims.UserID)

	if payload.Quiz.ID == uuid.Nil {
		payload.Quiz.ID = uuid.New()
	}

	if payload.Quiz.SubjectID == uuid.Nil {
		log.Warn("SubjectID inválido ou ausente")
		http.Error(w, "invalid subject_id", http.StatusBadRequest)
		return
	}

	for _, q := range payload.Questions {
		if q.ID == uuid.Nil {
			q.ID = uuid.New()
		}
		q.QuizID = payload.Quiz.ID
	}

	if err := h.service.CreateQuizWithQuestions(r.Context(), &payload.Quiz, payload.Questions); err != nil {
		log.WithError(err).Error("Erro ao criar quiz com perguntas")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	config.JSON(w, http.StatusCreated, map[string]interface{}{
		"quiz":      payload.Quiz,
		"questions": payload.Questions,
	})
}

func (h *Handler) DeleteQuiz(w http.ResponseWriter, r *http.Request) {
	log := config.WithContext(r.Context())

	quizID := chi.URLParam(r, "id")
	if quizID == "" {
		log.Warn("ID do quiz não fornecido")
		http.Error(w, "quiz id required", http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteQuiz(r.Context(), quizID); err != nil {
		log.WithError(err).Error("Erro ao deletar quiz")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	config.JSON(w, http.StatusOK, map[string]string{
		"message": "quiz deleted successfully",
	})
}

func (h *Handler) AddQuestion(w http.ResponseWriter, r *http.Request) {
	log := config.WithContext(r.Context())

	quizID := chi.URLParam(r, "id")
	if quizID == "" {
		log.Warn("ID do quiz não fornecido para adicionar pergunta")
		http.Error(w, "quiz id required", http.StatusBadRequest)
		return
	}

	var question QuizQuestion
	if err := json.NewDecoder(r.Body).Decode(&question); err != nil {
		log.WithError(err).Error("Corpo da requisição inválido para adicionar pergunta")
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if question.ID == uuid.Nil {
		question.ID = uuid.New()
	}

	if err := h.service.AddQuestionToQuiz(r.Context(), quizID, &question); err != nil {
		log.WithError(err).Error("Erro ao adicionar pergunta ao quiz")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	config.JSON(w, http.StatusCreated, map[string]interface{}{
		"message":  "question added successfully",
		"question": question,
	})
}

func (h *Handler) RemoveQuestion(w http.ResponseWriter, r *http.Request) {
	log := config.WithContext(r.Context())

	questionID := chi.URLParam(r, "questionID")
	if questionID == "" {
		log.Warn("ID da pergunta não fornecido")
		http.Error(w, "question id required", http.StatusBadRequest)
		return
	}

	if err := h.service.RemoveQuestion(r.Context(), questionID); err != nil {
		log.WithError(err).Error("Erro ao remover pergunta do quiz")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	config.JSON(w, http.StatusOK, map[string]string{
		"message": "question removed successfully",
	})
}

func (h *Handler) GetQuizWithQuestions(w http.ResponseWriter, r *http.Request) {
	log := config.WithContext(r.Context())

	quizID := chi.URLParam(r, "id")
	if quizID == "" {
		log.Warn("ID do quiz não fornecido")
		http.Error(w, "quiz id required", http.StatusBadRequest)
		return
	}

	quizWithQuestions, err := h.service.GetQuizWithQuestions(r.Context(), quizID)
	if err != nil {
		log.WithError(err).Error("Erro ao buscar quiz com perguntas")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if quizWithQuestions == nil {
		http.Error(w, "quiz not found", http.StatusNotFound)
		return
	}

	config.JSON(w, http.StatusOK, quizWithQuestions)
}

func (h *Handler) ListQuizzesByUser(w http.ResponseWriter, r *http.Request) {
	log := config.WithContext(r.Context())

	claims, err := auth.GetUserClaimsFromContext(r.Context())
	if err != nil {
		log.Warn("Usuário não autenticado para listar quizzes")
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	userID := claims.UserID
	if userID == "" {
		log.Warn("ID do usuário não fornecido")
		http.Error(w, "user id required", http.StatusBadRequest)
		return
	}

	quizzes, err := h.service.ListQuizzesByUser(r.Context(), userID)
	if err != nil {
		log.WithError(err).Error("Erro ao listar quizzes do usuário")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	config.JSON(w, http.StatusOK, quizzes)
}
