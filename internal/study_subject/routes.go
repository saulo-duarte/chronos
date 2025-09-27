package studysubject

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/saulo-duarte/chronos-lambda/internal/auth"
)

func Routes(h *Handler) http.Handler {
	r := chi.NewRouter()

	r.Use(auth.AuthMiddleware)

	r.Post("/", h.CreateStudySubject)
	r.Get("/", h.ListStudySubjects)
	r.Put("/{id}", h.UpdateStudySubject)
	r.Delete("/{id}", h.DeleteStudySubject)

	return r
}
