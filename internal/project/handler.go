package project

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/saulo-duarte/chronos-lambda/internal/auth"
	"github.com/saulo-duarte/chronos-lambda/internal/config"
)

type Handler struct {
	service ProjectService
}

func NewHandler(s ProjectService) *Handler {
	return &Handler{service: s}
}

func (h *Handler) CreateProject(w http.ResponseWriter, r *http.Request) {
	log := config.WithContext(r.Context())

	claims, err := auth.GetUserClaimsFromContext(r.Context())
	if err != nil {
		log.Warn("Usuário não autenticado para criar projeto")
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var payload Project
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		log.WithError(err).Error("Corpo da requisição inválido")
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	payload.UserID = uuid.MustParse(claims.UserID)

	project, err := h.service.CreateProject(r.Context(), &payload)
	if err != nil {
		log.WithError(err).Error("Erro ao criar projeto")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	config.JSON(w, http.StatusCreated, project)
}

func (h *Handler) GetProject(w http.ResponseWriter, r *http.Request) {
	log := config.WithContext(r.Context())

	projectID := chi.URLParam(r, "id")
	if projectID == "" {
		log.Warn("ID do projeto não fornecido")
		http.Error(w, "project id required", http.StatusBadRequest)
		return
	}

	project, err := h.service.GetProjectByID(r.Context(), projectID)
	if err != nil {
		if err == ErrProjectNotFound {
			http.Error(w, "project not found", http.StatusNotFound)
			return
		}
		log.WithError(err).Error("Erro ao buscar projeto")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	config.JSON(w, http.StatusOK, project)
}

func (h *Handler) ListProjects(w http.ResponseWriter, r *http.Request) {
	log := config.WithContext(r.Context())

	projects, err := h.service.ListProjectsByUser(r.Context())
	if err != nil {
		log.WithError(err).Error("Erro ao listar projetos")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	config.JSON(w, http.StatusOK, map[string]interface{}{
		"count":    len(projects),
		"projects": projects,
	})
}

func (h *Handler) UpdateProject(w http.ResponseWriter, r *http.Request) {
	log := config.WithContext(r.Context())

	projectID := chi.URLParam(r, "id")
	if projectID == "" {
		log.Warn("ID do projeto não fornecido")
		http.Error(w, "project id required", http.StatusBadRequest)
		return
	}

	var payload UpdateProjectDTO
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		log.WithError(err).Error("Corpo da requisição inválido")
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := payload.Validate(); err != nil {
		log.WithError(err).Warn("Payload inválido")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	project, err := h.service.UpdateProject(r.Context(), projectID, &payload)
	if err != nil {
		switch err {
		case ErrProjectNotFound:
			http.Error(w, "project not found", http.StatusNotFound)
		case ErrUnauthorized:
			http.Error(w, "unauthorized", http.StatusUnauthorized)
		default:
			log.WithError(err).Error("Erro ao atualizar projeto")
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	config.JSON(w, http.StatusOK, project)
}

func (h *Handler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	log := config.WithContext(r.Context())

	projectID := chi.URLParam(r, "id")
	if projectID == "" {
		log.Warn("ID do projeto não fornecido")
		http.Error(w, "project id required", http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteProject(r.Context(), projectID); err != nil {
		switch err {
		case ErrProjectNotFound:
			http.Error(w, "project not found", http.StatusNotFound)
		case ErrUnauthorized:
			http.Error(w, "unauthorized", http.StatusUnauthorized)
		default:
			log.WithError(err).Error("Erro ao deletar projeto")
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	config.JSON(w, http.StatusOK, map[string]string{
		"message": "project deleted successfully",
	})
}
