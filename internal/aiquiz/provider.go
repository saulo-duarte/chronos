package aiquiz

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/saulo-duarte/chronos-lambda/internal/config"
	"google.golang.org/genai"
)

type Provider interface {
	SendPrompt(ctx context.Context, system, user string) ([]Question, error)
}

type geminiProvider struct {
	client *genai.Client
}

func NewGeminiProvider(ctx context.Context) (Provider, error) {
	client, err := genai.NewClient(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar cliente Gemini: %w", err)
	}
	return &geminiProvider{client: client}, nil
}

func (p *geminiProvider) SendPrompt(ctx context.Context, system, user string) ([]Question, error) {
	log := config.WithContext(ctx)
	prompt := system + "\n\n" + user

	result, err := p.client.Models.GenerateContent(
		ctx,
		"gemini-2.0-flash",
		genai.Text(prompt),
		nil,
	)
	if err != nil {
		log.WithError(err).Error("falha ao gerar conteúdo do Gemini")
		return nil, fmt.Errorf("falha ao gerar conteúdo: %w", err)
	}

	raw := result.Text()
	log.Debugf("[AIQUIZ] Resposta bruta do Gemini:\n%s", raw)

	if raw == "" {
		return nil, errors.New("resposta vazia do modelo")
	}

	clean := strings.TrimSpace(raw)
	clean = strings.TrimPrefix(clean, "```json")
	clean = strings.TrimSuffix(clean, "```")
	clean = strings.Trim(clean, "`")

	var questions []Question
	if err := json.Unmarshal([]byte(clean), &questions); err != nil {
		log.WithError(err).Errorf("[AIQUIZ] Falha ao decodificar JSON. Conteúdo limpo:\n%s", clean)
		return nil, fmt.Errorf("falha ao decodificar JSON: %w", err)
	}

	log.Infof("[AIQUIZ] Geradas %d perguntas com sucesso", len(questions))
	return questions, nil
}
