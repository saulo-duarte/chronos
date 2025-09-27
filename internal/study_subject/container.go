package studysubject

import "gorm.io/gorm"

type StudySubjectContainer struct {
	Handler *Handler
}

func NewStudySubjectContainer(db *gorm.DB) *StudySubjectContainer {
	repo := NewRepository(db)
	service := NewService(repo)
	handler := NewHandler(service)

	return &StudySubjectContainer{
		Handler: handler,
	}
}
