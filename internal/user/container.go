package user

import "gorm.io/gorm"

type UserContainer struct {
	Handler *Handler
	Repo    UserRepository
}

func NewUserContainer(db *gorm.DB) *UserContainer {
	repo := NewRepository(db)
	service := NewService(repo)
	handler := NewHandler(service)

	return &UserContainer{
		Handler: handler,
		Repo:    repo,
	}
}
