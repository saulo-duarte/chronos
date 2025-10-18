package aiquiz

import (
	"encoding/json"
	"net/http"

	"github.com/saulo-duarte/chronos-lambda/internal/config"
)

type Handler struct {
	service Service
}

func NewHandler(s Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) GenerateQuestions(w http.ResponseWriter, r *http.Request) {
	log := config.WithContext(r.Context())
	var req QuestionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	questions, err := h.service.GenerateQuestions(r.Context(), req)
	if err != nil {
		http.Error(w, "failed to generate questions", http.StatusInternalServerError)
		log.WithError(err).Errorf("Failed to generate questions: %v", err)
		return
	}

	config.JSON(w, http.StatusCreated, questions)
}
