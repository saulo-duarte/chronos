package studytopic

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/saulo-duarte/chronos-lambda/internal/config"
	studysubject "github.com/saulo-duarte/chronos-lambda/internal/study_subject"
)

type Handler struct {
	service StudyTopicService
}

func NewHandler(s StudyTopicService) *Handler {
	return &Handler{service: s}
}

type createStudyTopicPayload struct {
	Name           string `json:"name"`
	Description    string `json:"description"`
	Position       int    `json:"position"`
	StudySubjectID string `json:"subject_id"`
}

func (h *Handler) CreateStudyTopic(w http.ResponseWriter, r *http.Request) {
	log := config.WithContext(r.Context())

	var payload createStudyTopicPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		log.WithError(err).Error("Invalid request body")
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	subjectID, err := uuid.Parse(payload.StudySubjectID)
	if err != nil {
		log.WithError(err).Warn("Invalid study subject ID format")
		http.Error(w, "invalid study subject id", http.StatusBadRequest)
		return
	}

	topicModel := &StudyTopic{
		Name:           payload.Name,
		Description:    payload.Description,
		Position:       payload.Position,
		StudySubjectID: subjectID,
	}

	topic, err := h.service.CreateStudyTopic(r.Context(), topicModel)
	if err != nil {
		switch {
		case errors.Is(err, ErrUnauthorized):
			http.Error(w, "unauthorized", http.StatusUnauthorized)
		case errors.Is(err, studysubject.ErrStudySubjectNotFound):
			http.Error(w, "study subject not found", http.StatusNotFound)
		case err.Error() == "já existe um tópico com esta posição neste assunto":
			http.Error(w, err.Error(), http.StatusConflict)
		default:
			log.WithError(err).Error("Error creating study topic")
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	config.JSON(w, http.StatusCreated, topic)
}

func (h *Handler) GetStudyTopic(w http.ResponseWriter, r *http.Request) {
	log := config.WithContext(r.Context())

	topicID := chi.URLParam(r, "id")
	if topicID == "" {
		log.Warn("Study topic ID not provided")
		http.Error(w, "study topic id required", http.StatusBadRequest)
		return
	}

	topic, err := h.service.GetStudyTopicByID(r.Context(), topicID)
	if err != nil {
		switch {
		case errors.Is(err, ErrUnauthorized):
			http.Error(w, "unauthorized", http.StatusUnauthorized)
		case errors.Is(err, ErrStudyTopicNotFound):
			http.Error(w, "study topic not found", http.StatusNotFound)
		default:
			log.WithError(err).Error("Error fetching study topic")
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	config.JSON(w, http.StatusOK, topic)
}

func (h *Handler) ListStudyTopics(w http.ResponseWriter, r *http.Request) {
	log := config.WithContext(r.Context())

	subjectID := chi.URLParam(r, "studySubjectId")
	if subjectID == "" {
		log.Warn("Study subject ID not provided for listing topics")
		http.Error(w, "study subject id required", http.StatusBadRequest)
		return
	}

	topics, err := h.service.ListStudyTopicsBySubject(r.Context(), subjectID)
	if err != nil {
		switch {
		case errors.Is(err, ErrUnauthorized):
			http.Error(w, "unauthorized", http.StatusUnauthorized)
		case errors.Is(err, studysubject.ErrStudySubjectNotFound):
			http.Error(w, "study subject not found", http.StatusNotFound)
		default:
			log.WithError(err).Error("Error listing study topics")
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	config.JSON(w, http.StatusOK, map[string]interface{}{
		"count":  len(topics),
		"topics": topics,
	})
}

func (h *Handler) UpdateStudyTopic(w http.ResponseWriter, r *http.Request) {
	log := config.WithContext(r.Context())

	topicID := chi.URLParam(r, "id")
	if topicID == "" {
		log.Warn("Study topic ID not provided")
		http.Error(w, "study topic id required", http.StatusBadRequest)
		return
	}

	var payload StudyTopic
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		log.WithError(err).Error("Invalid request body")
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	payload.ID = uuid.MustParse(topicID)

	topic, err := h.service.UpdateStudyTopic(r.Context(), &payload)
	if err != nil {
		switch {
		case errors.Is(err, ErrUnauthorized):
			http.Error(w, "unauthorized", http.StatusUnauthorized)
		case errors.Is(err, ErrStudyTopicNotFound):
			http.Error(w, "study topic not found", http.StatusNotFound)
		default:
			log.WithError(err).Error("Error updating study topic")
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	config.JSON(w, http.StatusOK, topic)
}

func (h *Handler) DeleteStudyTopic(w http.ResponseWriter, r *http.Request) {
	log := config.WithContext(r.Context())

	topicID := chi.URLParam(r, "id")
	if topicID == "" {
		log.Warn("Study topic ID not provided")
		http.Error(w, "study topic id required", http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteStudyTopic(r.Context(), topicID); err != nil {
		switch {
		case errors.Is(err, ErrUnauthorized):
			http.Error(w, "unauthorized", http.StatusUnauthorized)
		case errors.Is(err, ErrStudyTopicNotFound):
			http.Error(w, "study topic not found", http.StatusNotFound)
		default:
			log.WithError(err).Error("Error deleting study topic")
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	config.JSON(w, http.StatusOK, map[string]string{
		"message": "study topic deleted successfully",
	})
}
