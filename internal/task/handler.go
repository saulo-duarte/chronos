package task

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/saulo-duarte/chronos-lambda/internal/config"
)

type Handler struct {
	service TaskService
}

func NewHandler(s TaskService) *Handler {
	return &Handler{service: s}
}

func (h *Handler) CreateTask(w http.ResponseWriter, r *http.Request) {
	log := config.WithContext(r.Context())

	var payload Task
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		log.WithError(err).Error("Corpo da requisição inválido")
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	task, err := h.service.CreateTask(r.Context(), &payload)
	if err != nil {
		log.WithError(err).Error("Falha ao criar task")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	config.JSON(w, http.StatusCreated, task)
}

func (h *Handler) GetTask(w http.ResponseWriter, r *http.Request) {
	log := config.WithContext(r.Context())

	id := chi.URLParam(r, "taskID")

	task, err := h.service.FindByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrTaskNotFound) {
			http.Error(w, "task not found", http.StatusNotFound)
			return
		}
		log.WithError(err).Error("Erro ao buscar task")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	config.JSON(w, http.StatusOK, task)
}

func (h *Handler) ListTasksByUser(w http.ResponseWriter, r *http.Request) {
	log := config.WithContext(r.Context())

	tasks, err := h.service.FindAllByUser(r.Context())
	if err != nil {
		log.WithError(err).Error("Erro ao listar tasks por usuário")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	config.JSON(w, http.StatusOK, tasks)
}

func (h *Handler) ListTasksByProject(w http.ResponseWriter, r *http.Request) {
	log := config.WithContext(r.Context())

	projectID := chi.URLParam(r, "projectID")

	tasks, err := h.service.FindAllByProjectID(r.Context(), projectID)
	if err != nil {
		if errors.Is(err, ErrProjectNotFound) {
			http.Error(w, "project not found", http.StatusNotFound)
			return
		}
		log.WithError(err).Error("Erro ao listar tasks por projeto")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	config.JSON(w, http.StatusOK, tasks)
}

func (h *Handler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	log := config.WithContext(r.Context())

	id := chi.URLParam(r, "taskID")

	var payload Task
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		log.WithError(err).Error("Corpo da requisição inválido")
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	payload.ID, _ = uuid.Parse(id)

	task, err := h.service.UpdateTask(r.Context(), &payload)
	if err != nil {
		if errors.Is(err, ErrTaskNotFound) {
			http.Error(w, "task not found", http.StatusNotFound)
			return
		}
		log.WithError(err).Error("Erro ao atualizar task")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	config.JSON(w, http.StatusOK, task)
}

func (h *Handler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	log := config.WithContext(r.Context())

	id := chi.URLParam(r, "taskID")

	if err := h.service.DeleteByID(r.Context(), id); err != nil {
		if errors.Is(err, ErrTaskNotFound) {
			http.Error(w, "task not found", http.StatusNotFound)
			return
		}
		log.WithError(err).Error("Erro ao excluir task")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	config.JSON(w, http.StatusOK, map[string]string{
		"message": "task deleted successfully",
	})
}

func (h *Handler) ListTasksByStudyTopic(w http.ResponseWriter, r *http.Request) {
	log := config.WithContext(r.Context())

	studyTopicID := chi.URLParam(r, "studyTopicId")

	tasks, err := h.service.FindAllByTopicID(r.Context(), studyTopicID)
	if err != nil {
		if errors.Is(err, ErrStudyTopicNotFound) {
			http.Error(w, "study topic not found", http.StatusNotFound)
			return
		}
		log.WithError(err).Error("Erro ao listar tasks por tópico de estudo")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	config.JSON(w, http.StatusOK, tasks)
}
