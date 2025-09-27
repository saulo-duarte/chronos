package auth

import (
	"context"
	"errors"
)

type ClaimsFromContext struct {
	UserID string
	Role   string
}

var ErrNoAuthData = errors.New("no authentication data in context")

func GetUserClaimsFromContext(ctx context.Context) (*ClaimsFromContext, error) {
	userID, ok := ctx.Value(UserDataKeyID).(string)
	if !ok || userID == "" {
		return nil, ErrNoAuthData
	}

	role, ok := ctx.Value(UserDataKeyRole).(string)
	if !ok || role == "" {
		return nil, ErrNoAuthData
	}

	return &ClaimsFromContext{
		UserID: userID,
		Role:   role,
	}, nil
}
