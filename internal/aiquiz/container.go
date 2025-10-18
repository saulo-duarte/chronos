package aiquiz

import "context"

type AIQuizContainer struct {
	Handler *Handler
}

func NewAIQuizContainer() *AIQuizContainer {
	ctx := context.Background()
	provider, _ := NewGeminiProvider(ctx)
	service := NewService(provider)
	handler := NewHandler(service)

	return &AIQuizContainer{
		Handler: handler,
	}
}
