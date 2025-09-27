package user

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/saulo-duarte/chronos-lambda/internal/auth"
	"github.com/saulo-duarte/chronos-lambda/internal/config"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

type UserService interface {
	LoginWithGoogleUser(ctx context.Context, authResult *auth.AuthResult) (*User, string, string, error)
	Login(ctx context.Context, providerID string) (*User, string, string, error)
	RefreshToken(ctx context.Context, tokenString string) (string, error)
}

type userService struct {
	repo UserRepository
}

func NewService(repo UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) LoginWithGoogleUser(ctx context.Context, authResult *auth.AuthResult) (*User, string, string, error) {
	log := config.WithContext(ctx)

	providerID := authResult.ProviderID
	log.WithField("provider_id", providerID).Info("Payload do Google recebido com sucesso")

	user, err := s.repo.GetByProviderID(providerID)
	if err != nil && !errors.Is(err, ErrUserNotFound) {
		log.WithError(err).Error("Erro ao buscar usuário por provider ID")
		return nil, "", "", err
	}

	if user == nil {
		log.Info("Usuário não encontrado, criando novo usuário")
		user = &User{
			ID:                          uuid.New(),
			ProviderID:                  providerID,
			Username:                    authResult.Username,
			Email:                       authResult.Email,
			AvatarURL:                   authResult.Picture,
			Role:                        "USER",
			EncryptedGoogleAccessToken:  authResult.AccessToken,
			EncryptedGoogleRefreshToken: authResult.RefreshToken,
			CreatedAt:                   time.Now(),
			UpdatedAt:                   time.Now(),
		}
		if err := s.repo.Create(user); err != nil {
			log.WithError(err).Error("Falha ao criar novo usuário")
			return nil, "", "", err
		}
		log.WithField("user_id", user.ID).Info("Novo usuário criado com sucesso")
	} else {
		log.WithField("user_id", user.ID).Info("Usuário encontrado, atualizando informações")
		user.Username = authResult.Username
		user.Email = authResult.Email
		user.AvatarURL = authResult.Picture
		user.EncryptedGoogleAccessToken = authResult.AccessToken
		user.EncryptedGoogleRefreshToken = authResult.RefreshToken
		user.UpdatedAt = time.Now()
		if err := s.repo.Update(user); err != nil {
			log.WithError(err).Error("Falha ao atualizar usuário existente")
			return nil, "", "", err
		}
		log.WithField("user_id", user.ID).Info("Usuário atualizado com sucesso")
	}

	jwtToken, err := auth.GenerateJWT(user.ID.String(), user.Role, 24*time.Hour)
	if err != nil {
		log.WithError(err).Error("Falha ao gerar JWT")
		return nil, "", "", err
	}

	refreshToken, err := auth.GenerateJWT(user.ID.String(), user.Role, 14*24*time.Hour)
	if err != nil {
		log.WithError(err).Error("Falha ao gerar refresh token")
		return nil, "", "", err
	}

	log.WithField("user_id", user.ID).Info("Login via Google concluído com sucesso")

	return user, jwtToken, refreshToken, nil
}

func (s *userService) Login(ctx context.Context, providerID string) (*User, string, string, error) {
	log := config.WithContext(ctx)

	user, err := s.repo.GetByProviderID(providerID)
	if err != nil {
		log.WithError(err).Error("Erro ao buscar usuário por provider ID")
		return nil, "", "", ErrUserNotFound
	}
	if user == nil {
		log.WithField("provider_id", providerID).Warn("Usuário não encontrado para login")
		return nil, "", "", ErrUserNotFound
	}

	jwtToken, err := auth.GenerateJWT(user.ID.String(), user.Role, 24*time.Hour)
	if err != nil {
		log.WithError(err).Error("Falha ao gerar JWT")
		return nil, "", "", err
	}

	refreshToken, err := auth.GenerateJWT(user.ID.String(), user.Role, 14*24*time.Hour)
	if err != nil {
		log.WithError(err).Error("Falha ao gerar refresh token")
		return nil, "", "", err
	}

	log.WithField("user_id", user.ID).Info("Login de usuário realizado com sucesso")
	return user, jwtToken, refreshToken, nil
}

func (s *userService) RefreshToken(ctx context.Context, tokenString string) (string, error) {
	log := config.WithContext(ctx)

	claims, err := auth.ValidateJWT(tokenString)
	if err != nil {
		log.WithError(err).Warn("Refresh token inválido")
		return "", errors.New("invalid refresh token")
	}

	user, err := s.repo.GetByID(claims.UserID)
	if err != nil {
		log.WithError(err).Error("Erro ao buscar usuário para refresh token")
		return "", err
	}
	if user == nil {
		log.WithField("user_id", claims.UserID).Warn("Usuário não encontrado para refresh token")
		return "", ErrUserNotFound
	}

	newJWT, err := auth.GenerateJWT(user.ID.String(), user.Role, 24*time.Hour)
	if err != nil {
		log.WithError(err).Error("Falha ao gerar novo JWT")
		return "", err
	}

	log.WithField("user_id", user.ID).Info("JWT atualizado com sucesso")
	return newJWT, nil
}
