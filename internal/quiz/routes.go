package quiz

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/saulo-duarte/chronos-lambda/internal/auth"
)

func Routes(h *Handler) http.Handler {
	r := chi.NewRouter()

	r.Use(auth.AuthMiddleware)

	r.Post("/", h.CreateQuiz)
	r.Get("/", h.ListQuizzesByUser)
	r.Get("/{id}", h.GetQuizWithQuestions)
	r.Delete("/{id}", h.DeleteQuiz)
	r.Post("/{id}/questions", h.AddQuestion)
	r.Delete("/questions/{questionID}", h.RemoveQuestion)
	return r
}
