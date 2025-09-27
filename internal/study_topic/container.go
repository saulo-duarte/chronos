package studytopic

import (
	studysubject "github.com/saulo-duarte/chronos-lambda/internal/study_subject"
	"gorm.io/gorm"
)

type StudyTopicContainer struct {
	Handler *Handler
	Repo    StudyTopicRepository
}

func NewStudyTopicContainer(db *gorm.DB) *StudyTopicContainer {
	studySubjectRepo := studysubject.NewRepository(db)
	repo := NewRepository(db)
	service := NewService(repo, studySubjectRepo)
	handler := NewHandler(service)

	return &StudyTopicContainer{
		Handler: handler,
		Repo:    repo,
	}
}
