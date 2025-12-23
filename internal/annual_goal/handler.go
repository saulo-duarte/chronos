package annual_goal

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/saulo-duarte/chronos-lambda/internal/auth"
	"github.com/saulo-duarte/chronos-lambda/internal/config"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	log := config.WithContext(r.Context())

	claims, err := auth.GetUserClaimsFromContext(r.Context())
	if err != nil {
		log.Warn("User not authenticated")
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var dto CreateAnnualGoalDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		log.WithError(err).Error("Invalid request body")
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	userID := uuid.MustParse(claims.UserID)
	response, err := h.service.Create(userID, dto)
	if err != nil {
		log.WithError(err).Error("Failed to create annual goal")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	config.JSON(w, http.StatusCreated, response)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	log := config.WithContext(r.Context())

	claims, err := auth.GetUserClaimsFromContext(r.Context())
	if err != nil {
		log.Warn("User not authenticated")
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	userID := uuid.MustParse(claims.UserID)
	responses, err := h.service.List(userID)
	if err != nil {
		log.WithError(err).Error("Failed to list annual goals")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	config.JSON(w, http.StatusOK, responses)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	log := config.WithContext(r.Context())

	claims, err := auth.GetUserClaimsFromContext(r.Context())
	if err != nil {
		log.Warn("User not authenticated")
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		http.Error(w, "id required", http.StatusBadRequest)
		return
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var dto UpdateAnnualGoalDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		log.WithError(err).Error("Invalid request body")
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	userID := uuid.MustParse(claims.UserID)
	response, err := h.service.Update(id, userID, dto)
	if err != nil {
		if err.Error() == "unauthorized" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		log.WithError(err).Error("Failed to update annual goal")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	config.JSON(w, http.StatusOK, response)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	log := config.WithContext(r.Context())

	claims, err := auth.GetUserClaimsFromContext(r.Context())
	if err != nil {
		log.Warn("User not authenticated")
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		http.Error(w, "id required", http.StatusBadRequest)
		return
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	userID := uuid.MustParse(claims.UserID)
	if err := h.service.Delete(id, userID); err != nil {
		if err.Error() == "unauthorized" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		log.WithError(err).Error("Failed to delete annual goal")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
