package annual_goal

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func Routes(h *Handler) http.Handler {
	r := chi.NewRouter()

	r.Post("/", h.Create)
	r.Get("/", h.List)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)

	return r
}
