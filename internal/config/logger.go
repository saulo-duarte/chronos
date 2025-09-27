package config

import (
	"context"
	"os"

	"github.com/sirupsen/logrus"
)

var Logger *logrus.Logger

func Init() {
	Logger = logrus.New()

	Logger.SetFormatter(&logrus.JSONFormatter{})

	if level, err := logrus.ParseLevel(os.Getenv("LOG_LEVEL")); err == nil {
		Logger.SetLevel(level)
	} else {
		Logger.SetLevel(logrus.InfoLevel)
	}
}

func WithContext(ctx context.Context) *logrus.Entry {
	requestID, _ := ctx.Value("requestID").(string)
	return Logger.WithFields(logrus.Fields{
		"lambda_request_id": requestID,
	})
}
