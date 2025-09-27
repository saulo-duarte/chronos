package task

import (
	"github.com/go-chi/chi/v5"
)

func Routes(h *Handler) chi.Router {
	r := chi.NewRouter()

	r.Post("/", h.CreateTask)
	r.Get("/{taskID}", h.GetTask)
	r.Get("/", h.ListTasksByUser)
	r.Get("/project/{projectID}", h.ListTasksByProject)
	r.Put("/{taskID}", h.UpdateTask)
	r.Delete("/{taskID}", h.DeleteTask)

	return r
}
