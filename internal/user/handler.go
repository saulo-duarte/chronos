package user

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/saulo-duarte/chronos-lambda/internal/auth"
	"github.com/saulo-duarte/chronos-lambda/internal/config"
)

var FRONTEND_URL = os.Getenv("FRONTEND_URL")
var ENV = os.Getenv("ENV")
var API_DOMAIN = os.Getenv("API_DOMAIN")

type Handler struct {
	service UserService
}

func NewHandler(s UserService) *Handler {
	return &Handler{service: s}
}

func newCookie(name, value string, maxAge int) *http.Cookie {
	secure := ENV == "prod"
	c := &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   maxAge,
		SameSite: http.SameSiteNoneMode,
		Secure:   secure,
		Domain:   API_DOMAIN,
	}
	return c
}

func (h *Handler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	log := config.WithContext(r.Context())

	var payload auth.AuthResult
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		log.WithError(err).Error("Corpo da requisição inválido")
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if payload.ProviderID == "" || payload.Email == "" {
		http.Error(w, "providerID and email are required", http.StatusBadRequest)
		return
	}

	user, jwtToken, refreshToken, err := h.service.LoginWithGoogleUser(r.Context(), &payload)
	if err != nil {
		log.WithError(err).Error("Falha no login via Google")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	jwtCookie := newCookie(auth.JWT_COOKIE_NAME, jwtToken, int((24 * time.Hour).Seconds()))
	refreshCookie := newCookie(auth.REFRESH_TOKEN_COOKIE_NAME, refreshToken, int((14 * 24 * time.Hour).Seconds()))

	http.SetCookie(w, jwtCookie)
	http.SetCookie(w, refreshCookie)

	config.JSON(w, http.StatusOK, map[string]any{
		"user":    user.ToResponse(),
		"message": "Login successful",
	})
}

func (h *Handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	log := config.WithContext(r.Context())

	cookie, err := r.Cookie(auth.REFRESH_TOKEN_COOKIE_NAME)
	if err != nil {
		config.JSON(w, http.StatusUnauthorized, map[string]string{
			"error": "refresh token required",
		})
		return
	}

	newJWT, err := h.service.RefreshToken(r.Context(), cookie.Value)
	if err != nil {
		log.WithError(err).Error("Falha ao atualizar o token")
		config.JSON(w, http.StatusUnauthorized, map[string]string{
			"error": "failed to refresh token",
		})
		return
	}

	http.SetCookie(w, newCookie(auth.JWT_COOKIE_NAME, newJWT, int((24*time.Hour).Seconds())))

	config.JSON(w, http.StatusOK, map[string]string{
		"message": "token refreshed successfully",
	})
}

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	log := config.WithContext(r.Context())

	claims, err := auth.GetUserClaimsFromContext(r.Context())
	if err != nil {
		log.WithError(err).Error("Falha ao obter claims do contexto")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	userID := claims.UserID

	if userID == "" {
		http.Error(w, "userID is required", http.StatusBadRequest)
		return
	}
	user, err := h.service.GetByID(r.Context(), userID)
	if err != nil {
		log.WithError(err).Error("Erro ao buscar usuário")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if user == nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	config.JSON(w, http.StatusOK, user.ToResponse())
}
