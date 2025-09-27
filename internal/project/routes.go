package project

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/saulo-duarte/chronos-lambda/internal/auth"
)

func Routes(h *Handler) http.Handler {
	r := chi.NewRouter()

	r.Use(auth.AuthMiddleware)

	r.Post("/", h.CreateProject)
	r.Get("/", h.ListProjects)
	r.Get("/{id}", h.GetProject)
	r.Put("/{id}", h.UpdateProject)
	r.Delete("/{id}", h.DeleteProject)

	return r
}
