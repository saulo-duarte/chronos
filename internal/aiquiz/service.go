package aiquiz

import (
	"context"
)

type Service interface {
	GenerateQuestions(ctx context.Context, req QuestionRequest) ([]Question, error)
}

type service struct {
	provider Provider
}

func NewService(provider Provider) Service {
	return &service{provider: provider}
}

func (s *service) GenerateQuestions(ctx context.Context, req QuestionRequest) ([]Question, error) {
	system := systemPrompt
	user := BuildUserPrompt(req)

	return s.provider.SendPrompt(ctx, system, user)
}
