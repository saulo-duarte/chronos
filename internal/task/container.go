package task

import (
	googlecalendar "github.com/saulo-duarte/chronos-lambda/internal/google_calendar"
	"github.com/saulo-duarte/chronos-lambda/internal/project"
	studytopic "github.com/saulo-duarte/chronos-lambda/internal/study_topic"
	"github.com/saulo-duarte/chronos-lambda/internal/user"
	"gorm.io/gorm"
)

type TaskContainer struct {
	Handler *Handler
}

func NewTaskContainer(
	db *gorm.DB,
	projectService project.ProjectService,
	studyTopicRepo studytopic.StudyTopicRepository,
	userRepository user.UserRepository,
	calendarManager googlecalendar.CalendarManager,
) *TaskContainer {
	repo := NewRepository(db)
	service := NewService(repo, projectService, userRepository, studyTopicRepo, calendarManager)
	handler := NewHandler(service)

	return &TaskContainer{
		Handler: handler,
	}
}
