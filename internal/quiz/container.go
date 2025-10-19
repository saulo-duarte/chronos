package quiz

import "gorm.io/gorm"

type QuizContainer struct {
	Handler *Handler
}

func NewQuizContainer(db *gorm.DB) *QuizContainer {
	repo := NewRepository(db)
	service := NewService(db, repo)
	handler := NewHandler(service)

	return &QuizContainer{
		Handler: handler,
	}
}
