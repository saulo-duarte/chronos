package project

import "gorm.io/gorm"

type ProjectContainer struct {
	Handler *Handler
	Service ProjectService
}

func NewProjectContainer(db *gorm.DB) *ProjectContainer {
	repo := NewRepository(db)
	service := NewService(repo)
	handler := NewHandler(service)

	return &ProjectContainer{
		Handler: handler,
		Service: service,
	}
}
