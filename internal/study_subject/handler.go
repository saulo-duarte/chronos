package studysubject

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/saulo-duarte/chronos-lambda/internal/auth"
	"github.com/saulo-duarte/chronos-lambda/internal/config"
)

type Handler struct {
	service StudySubjectService
}

func NewHandler(s StudySubjectService) *Handler {
	return &Handler{service: s}
}

func (h *Handler) CreateStudySubject(w http.ResponseWriter, r *http.Request) {
	log := config.WithContext(r.Context())

	var payload StudySubject
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		log.WithError(err).Error("Invalid request body")
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	subject, err := h.service.CreateStudySubject(r.Context(), &payload)
	if err != nil {
		if err == ErrUnauthorized {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		log.WithError(err).Error("Error creating study subject")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	config.JSON(w, http.StatusCreated, subject)
}

func (h *Handler) ListStudySubjects(w http.ResponseWriter, r *http.Request) {
	log := config.WithContext(r.Context())

	claims, err := auth.GetUserClaimsFromContext(r.Context())
	if err != nil {
		log.WithError(err).Warn("Attempt to list study subjects without authentication")
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	subjects, err := h.service.ListStudySubjectsByUser(r.Context(), claims.UserID)
	if err != nil {
		if err == ErrUnauthorized {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		log.WithError(err).Error("Error listing study subjects")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	config.JSON(w, http.StatusOK, map[string]interface{}{
		"count":    len(subjects),
		"subjects": subjects,
	})
}

func (h *Handler) UpdateStudySubject(w http.ResponseWriter, r *http.Request) {
	log := config.WithContext(r.Context())

	subjectID := chi.URLParam(r, "id")
	if subjectID == "" {
		log.Warn("Study subject ID not provided")
		http.Error(w, "study subject id required", http.StatusBadRequest)
		return
	}

	var payload StudySubject
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		log.WithError(err).Error("Invalid request body")
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	payload.ID = uuid.MustParse(subjectID)

	subject, err := h.service.UpdateStudySubject(r.Context(), &payload)
	if err != nil {
		switch err {
		case ErrStudySubjectNotFound:
			http.Error(w, "study subject not found", http.StatusNotFound)
		case ErrUnauthorized:
			http.Error(w, "unauthorized", http.StatusUnauthorized)
		default:
			log.WithError(err).Error("Error updating study subject")
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	config.JSON(w, http.StatusOK, subject)
}

func (h *Handler) DeleteStudySubject(w http.ResponseWriter, r *http.Request) {
	log := config.WithContext(r.Context())

	subjectID := chi.URLParam(r, "id")
	if subjectID == "" {
		log.Warn("Study subject ID not provided")
		http.Error(w, "study subject id required", http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteStudySubject(r.Context(), subjectID); err != nil {
		switch err {
		case ErrStudySubjectNotFound:
			http.Error(w, "study subject not found", http.StatusNotFound)
		case ErrUnauthorized:
			http.Error(w, "unauthorized", http.StatusUnauthorized)
		default:
			log.WithError(err).Error("Error deleting study subject")
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	config.JSON(w, http.StatusOK, map[string]string{
		"message": "study subject deleted successfully",
	})
}
