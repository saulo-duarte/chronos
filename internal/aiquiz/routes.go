package aiquiz

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func Routes(h *Handler) http.Handler {
	r := chi.NewRouter()

	r.Post("/", h.GenerateQuestions)
	return r
}
