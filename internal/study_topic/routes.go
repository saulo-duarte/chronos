package studytopic

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/saulo-duarte/chronos-lambda/internal/auth"
)

func Routes(h *Handler) http.Handler {
	r := chi.NewRouter()

	r.Use(auth.AuthMiddleware)

	r.Post("/", h.CreateStudyTopic)
	r.Get("/{id}", h.ListStudyTopics)
	r.Put("/{id}", h.UpdateStudyTopic)
	r.Delete("/{id}", h.DeleteStudyTopic)
	r.Get("/{id}", h.GetStudyTopic)

	return r
}
