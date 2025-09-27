package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	chiadapter "github.com/awslabs/aws-lambda-go-api-proxy/chi"
	"github.com/go-chi/chi/v5"

	"github.com/saulo-duarte/chronos-lambda/internal/container"
	"github.com/saulo-duarte/chronos-lambda/internal/router"
)

var chiLambda *chiadapter.ChiLambdaV2
var chiRouter *chi.Mux

func init() {
	c := container.New()

	r := router.New(router.RouterConfig{
		UserHandler:         c.UserContainer.Handler,
		ProjectHandler:      c.ProjectContainer.Handler,
		TaskHandler:         c.TaskContainer.Handler,
		StudySubjectHandler: c.StudySubjectContainer.Handler,
		StudyTopicHandler:   c.StudyTopicContainer.Handler,
	})

	chiRouter = r.(*chi.Mux)

	chiLambda = chiadapter.NewV2(chiRouter)
}

func Handler(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	resp, err := chiLambda.ProxyWithContextV2(ctx, req)
	if err != nil {
		log.Printf("ERROR: ProxyWithContextV2 returned an error: %v\n", err)
	}
	return resp, err
}

func main() {
	runMode := os.Getenv("RUN_MODE")

	if runMode == "local" {
		log.Println("Iniciando servidor HTTP local em :3000")
		if err := http.ListenAndServe(":3000", chiRouter); err != nil {
			log.Fatalf("Falha ao iniciar servidor local: %v", err)
		}
	} else {
		lambda.Start(Handler)
	}
}
