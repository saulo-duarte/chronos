package container

import (
	"context"
	"log"
	"os"

	"github.com/saulo-duarte/chronos-lambda/internal/aiquiz"
	"github.com/saulo-duarte/chronos-lambda/internal/auth"
	"github.com/saulo-duarte/chronos-lambda/internal/config"
	googlecalendar "github.com/saulo-duarte/chronos-lambda/internal/google_calendar"
	"github.com/saulo-duarte/chronos-lambda/internal/project"
	"github.com/saulo-duarte/chronos-lambda/internal/quiz"
	studysubject "github.com/saulo-duarte/chronos-lambda/internal/study_subject"
	studytopic "github.com/saulo-duarte/chronos-lambda/internal/study_topic"
	"github.com/saulo-duarte/chronos-lambda/internal/task"
	"github.com/saulo-duarte/chronos-lambda/internal/user"
)

type Container struct {
	UserContainer           *user.UserContainer
	ProjectContainer        *project.ProjectContainer
	TaskContainer           *task.TaskContainer
	StudySubjectContainer   *studysubject.StudySubjectContainer
	StudyTopicContainer     *studytopic.StudyTopicContainer
	GoogleCalendarContainer *googlecalendar.GoogleCalendarContainer
	AIQuizContainer         *aiquiz.AIQuizContainer
	QuizContainer           *quiz.QuizContainer
}

func New() *Container {
	config.Init()
	auth.Init()
	config.InitCrypto()

	dsn := os.Getenv("DATABASE_DSN")
	if err := config.Connect(context.Background(), dsn); err != nil {
		log.Fatalf("failed to connect to DB: %v", err)
	}

	userContainer := user.NewUserContainer(config.DB)
	projectContainer := project.NewProjectContainer(config.DB)
	studySubjectContainer := studysubject.NewStudySubjectContainer(config.DB)
	studyTopicContainer := studytopic.NewStudyTopicContainer(config.DB)
	calendarContainer := googlecalendar.NewGoogleCalendarContainer(userContainer.Repo)
	aiQuizContainer := aiquiz.NewAIQuizContainer()
	quizContainer := quiz.NewQuizContainer(config.DB)

	taskContainer := task.NewTaskContainer(
		config.DB,
		projectContainer.Service,
		studyTopicContainer.Repo,
		userContainer.Repo,
		calendarContainer.CalendarManager,
	)

	return &Container{
		UserContainer:         userContainer,
		ProjectContainer:      projectContainer,
		TaskContainer:         taskContainer,
		StudySubjectContainer: studySubjectContainer,
		StudyTopicContainer:   studyTopicContainer,
		AIQuizContainer:       aiQuizContainer,
		QuizContainer:         quizContainer,
	}
}
