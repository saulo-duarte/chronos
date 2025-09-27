package task

import (
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
) *TaskContainer {
	repo := NewRepository(db)
	service := NewService(repo, projectService, userRepository, studyTopicRepo)
	handler := NewHandler(service)

	return &TaskContainer{
		Handler: handler,
	}
}
