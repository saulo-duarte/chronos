package annual_goal

import "gorm.io/gorm"

type Container struct {
	Handler *Handler
	Service Service
}

func NewContainer(db *gorm.DB) *Container {
	repo := NewRepository(db)
	service := NewService(repo)
	handler := NewHandler(service)

	return &Container{
		Handler: handler,
		Service: service,
	}
}
