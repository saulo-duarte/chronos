package auth

import (
	"encoding/json"
	"net/http"

	"github.com/saulo-duarte/chronos-lambda/internal/config"
	"github.com/sirupsen/logrus"
)

type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	VerifiedEmail bool   `json:"verified_email"`
}

type AuthResult struct {
	ProviderID   string
	Username     string
	Email        string
	Picture      string
	AccessToken  string
	RefreshToken string
}

type CallbackPayload struct {
	User   GoogleUserInfo   `json:"user"`
	Tokens GoogleTokenReply `json:"tokens"`
}

type GoogleTokenReply struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func GoogleCodeHandler(w http.ResponseWriter, r *http.Request) {
	log := config.WithContext(r.Context())

	var payload CallbackPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		log.WithError(err).Error("Payload inválido recebido do frontend")
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	encryptedAccessToken := ""
	encryptedRefreshToken := ""

	if payload.Tokens.AccessToken != "" {
		if enc, err := config.Encrypt(payload.Tokens.AccessToken); err == nil {
			encryptedAccessToken = enc
		}
	}

	if payload.Tokens.RefreshToken != "" {
		if enc, err := config.Encrypt(payload.Tokens.RefreshToken); err == nil {
			encryptedRefreshToken = enc
		}
	}

	authResult := &AuthResult{
		ProviderID:   payload.User.ID,
		Username:     payload.User.Name,
		Email:        payload.User.Email,
		Picture:      payload.User.Picture,
		AccessToken:  encryptedAccessToken,
		RefreshToken: encryptedRefreshToken,
	}

	log.WithFields(logrus.Fields{
		"email":   payload.User.Email,
		"user_id": payload.User.ID,
	}).Info("Usuário autenticado com sucesso via Google")

	json.NewEncoder(w).Encode(map[string]any{
		"user":    authResult,
		"message": "Google login successful",
	})
}
