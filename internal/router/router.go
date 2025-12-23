package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/saulo-duarte/chronos-lambda/internal/aiquiz"
	"github.com/saulo-duarte/chronos-lambda/internal/annual_goal"
	"github.com/saulo-duarte/chronos-lambda/internal/auth"
	"github.com/saulo-duarte/chronos-lambda/internal/middlewares"
	"github.com/saulo-duarte/chronos-lambda/internal/project"
	"github.com/saulo-duarte/chronos-lambda/internal/quiz"
	studysubject "github.com/saulo-duarte/chronos-lambda/internal/study_subject"
	studytopic "github.com/saulo-duarte/chronos-lambda/internal/study_topic"
	"github.com/saulo-duarte/chronos-lambda/internal/task"
	"github.com/saulo-duarte/chronos-lambda/internal/user"
)

type RouterConfig struct {
	UserHandler         *user.Handler
	ProjectHandler      *project.Handler
	TaskHandler         *task.Handler
	StudySubjectHandler *studysubject.Handler
	StudyTopicHandler   *studytopic.Handler
	AIQuizHandler       *aiquiz.Handler
	QuizHandler         *quiz.Handler
	AnnualGoalHandler   *annual_goal.Handler
}

func New(cfg RouterConfig) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middlewares.CorsMiddleware)

	r.Get("/swagger/*", httpSwagger.WrapHandler)

	r.Route("/ai-quiz", func(r chi.Router) {
		r.Mount("/", aiquiz.Routes(cfg.AIQuizHandler))
	})

	r.Route("/auth", func(r chi.Router) {
		r.Post("/login", cfg.UserHandler.GoogleLogin)
		r.Post("/refresh", cfg.UserHandler.RefreshToken)
		r.Post("/logout", auth.NewHandler().Logout)
	})

	r.Group(func(r chi.Router) {
		r.Use(auth.AuthMiddleware)

		r.Mount("/projects", project.Routes(cfg.ProjectHandler))
		r.Mount("/tasks", task.Routes(cfg.TaskHandler))
		r.Mount("/study-subjects", studysubject.Routes(cfg.StudySubjectHandler))
		r.Mount("/study-topics", studytopic.Routes(cfg.StudyTopicHandler))
		r.Mount("/users", user.Routes(cfg.UserHandler))
		r.Mount("/quizzes", quiz.Routes(cfg.QuizHandler))
		r.Mount("/annual-goals", annual_goal.Routes(cfg.AnnualGoalHandler))

		r.Get("/study-subjects/{studySubjectId}/topics", cfg.StudyTopicHandler.ListStudyTopics)
		r.Get("/study-topics/{studyTopicId}/tasks", cfg.TaskHandler.ListTasksByStudyTopic)
	})
	return r
}
