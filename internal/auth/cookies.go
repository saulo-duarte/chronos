package auth

import (
	"fmt"
	"time"
)

const (
	JWT_COOKIE_NAME           = "jwt"
	REFRESH_TOKEN_COOKIE_NAME = "refresh_token"
)

func NewJWTCookie(token string, duration time.Duration) string {
	return fmt.Sprintf("%s=%s; Path=/; Expires=%s; HttpOnly; SameSite=None; Secure", JWT_COOKIE_NAME, token, time.Now().Add(duration).UTC().Format(time.RFC1123))
}

func NewRefreshTokenCookie(token string, duration time.Duration) string {
	return fmt.Sprintf("%s=%s; Path=/; Expires=%s; HttpOnly; SameSite=None; Secure", REFRESH_TOKEN_COOKIE_NAME, token, time.Now().Add(duration).UTC().Format(time.RFC1123))
}
