package auth

import (
	"net/http"

	"github.com/saulo-duarte/chronos-lambda/internal/config"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "jwt",
		Value:    "",
		Path:     "/",
		Domain:   ".chronosapp.site",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
	})

	config.JSON(w, http.StatusOK, map[string]string{
		"message": "logout successful",
	})
}
