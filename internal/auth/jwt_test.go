package auth_test

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/saulo-duarte/chronos-lambda/internal/auth"
)

const testSecret = "uma-chave-secreta-para-testes-segura-e-longa"
const testUserID = "user-123"
const testRole = "admin"

var jwtSecret []byte

func TestInit(t *testing.T) {
	t.Run("MissingSecret", func(t *testing.T) {
		os.Unsetenv("JWT_SECRET")

		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Init() deveria ter causado pânico quando JWT_SECRET está vazio, mas não o fez.")
			}
		}()

		auth.Init()
	})

	t.Run("ValidSecret", func(t *testing.T) {
		os.Setenv("JWT_SECRET", testSecret)
		auth.Init()
	})
}

func TestGenerateAndValidateJWT(t *testing.T) {
	os.Setenv("JWT_SECRET", testSecret)
	auth.Init()

	t.Run("ValidToken", func(t *testing.T) {
		duration := time.Minute * 5

		tokenStr, err := auth.GenerateJWT(testUserID, testRole, duration)
		if err != nil {
			t.Fatalf("GenerateJWT falhou: %v", err)
		}

		claims, err := auth.ValidateJWT(tokenStr)
		if err != nil {
			t.Fatalf("ValidateJWT falhou inesperadamente: %v", err)
		}

		if claims.UserID != testUserID {
			t.Errorf("UserID incorreto. Esperado: %s, Recebido: %s", testUserID, claims.UserID)
		}
		if claims.Role != testRole {
			t.Errorf("Role incorreto. Esperado: %s, Recebido: %s", testRole, claims.Role)
		}
	})

	t.Run("ExpiredToken", func(t *testing.T) {
		duration := -time.Second * 1

		tokenStr, err := auth.GenerateJWT(testUserID, testRole, duration)
		if err != nil {
			t.Fatalf("GenerateJWT falhou: %v", err)
		}

		time.Sleep(time.Second * 2)

		_, err = auth.ValidateJWT(tokenStr)

		if err == nil {
			t.Fatal("ValidateJWT deveria ter falhado com token expirado, mas passou.")
		}
		if !errors.Is(err, jwt.ErrTokenExpired) {
			t.Errorf("Erro incorreto retornado para token expirado. Esperado: %v, Recebido: %v", jwt.ErrTokenExpired, err)
		}
	})

	t.Run("InvalidSignature", func(t *testing.T) {
		tokenStr, err := auth.GenerateJWT(testUserID, testRole, time.Minute)
		if err != nil {
			t.Fatalf("GenerateJWT falhou: %v", err)
		}

		originalSecret := jwtSecret
		jwtSecret = []byte("chave-secreta-falsa-diferente")

		_, err = auth.ValidateJWT(tokenStr)

		jwtSecret = originalSecret

		if err == nil {
			t.Fatal("ValidateJWT deveria ter falhado com assinatura inválida, mas passou.")
		}

		if err.Error() != "token is invalid: signature is invalid" {
			t.Errorf("Erro incorreto para assinatura inválida: %v", err)
		}
	})
}
