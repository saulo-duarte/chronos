package config

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
)

func APIGateway(ctx context.Context, statusCode int, body interface{}, headers map[string]string) (events.APIGatewayProxyResponse, error) {
	log := WithContext(ctx)

	jsonBody, err := json.Marshal(body)
	if err != nil {
		log.WithError(err).Error("Falha ao serializar o corpo da resposta")
		return APIGateway(ctx, http.StatusInternalServerError, map[string]string{"error": "failed to marshal response"}, headers)
	}

	log.WithFields(map[string]interface{}{
		"status_code": statusCode,
		"body":        string(jsonBody),
	}).Info("Gerando resposta para API Gateway")

	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Body:       string(jsonBody),
		Headers:    headers,
	}, nil
}

func NotFound(ctx context.Context, headers map[string]string) (events.APIGatewayProxyResponse, error) {
	return APIGateway(ctx, http.StatusNotFound, map[string]string{"error": "route not found"}, headers)
}

func BadRequest(ctx context.Context, msg string, headers map[string]string) (events.APIGatewayProxyResponse, error) {
	return APIGateway(ctx, http.StatusBadRequest, map[string]string{"error": msg}, headers)
}

func Unauthorized(ctx context.Context, msg string, headers map[string]string) (events.APIGatewayProxyResponse, error) {
	return APIGateway(ctx, http.StatusUnauthorized, map[string]string{"error": msg}, headers)
}

func InternalError(ctx context.Context, headers map[string]string) (events.APIGatewayProxyResponse, error) {
	return APIGateway(ctx, http.StatusInternalServerError, map[string]string{"error": "internal server error"}, headers)
}

func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
