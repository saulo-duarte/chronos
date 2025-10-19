package quiz

import (
	"errors"

	"gorm.io/gorm"
)

type QuizRepository interface {
	Create(q *Quiz) error
	GetByID(id string) (*Quiz, error)
	GetBySubjectAndUser(subjectID, userID string) ([]*Quiz, error)
	Delete(id string) error

	AddQuestions(questions []*QuizQuestion) error
	ListQuestionsByQuiz(quizID string) ([]*QuizQuestion, error)
	DeleteQuestion(id string) error
	ListQuizzesByUser(userID string) ([]*Quiz, error)
}

type quizRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) QuizRepository {
	return &quizRepository{db: db}
}

func (r *quizRepository) Create(q *Quiz) error {
	return r.db.Create(q).Error
}

func (r *quizRepository) GetByID(id string) (*Quiz, error) {
	var quiz Quiz
	if err := r.db.Preload("Questions").First(&quiz, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &quiz, nil
}

func (r *quizRepository) GetBySubjectAndUser(subjectID, userID string) ([]*Quiz, error) {
	var quizzes []*Quiz
	if err := r.db.
		Where("subject_id = ? AND user_id = ?", subjectID, userID).
		Order("created_at DESC").
		Find(&quizzes).Error; err != nil {
		return nil, err
	}
	return quizzes, nil
}

func (r *quizRepository) Delete(id string) error {
	return r.db.Delete(&Quiz{}, "id = ?", id).Error
}

func (r *quizRepository) AddQuestions(questions []*QuizQuestion) error {
	if len(questions) == 0 {
		return nil
	}
	return r.db.Create(&questions).Error
}

func (r *quizRepository) ListQuestionsByQuiz(quizID string) ([]*QuizQuestion, error) {
	var questions []*QuizQuestion
	if err := r.db.
		Where("quiz_id = ?", quizID).
		Order("order_index ASC").
		Find(&questions).Error; err != nil {
		return nil, err
	}
	return questions, nil
}

func (r *quizRepository) DeleteQuestion(id string) error {
	return r.db.Delete(&QuizQuestion{}, "id = ?", id).Error
}

func (r *quizRepository) ListQuizzesByUser(userID string) ([]*Quiz, error) {
	var quizzes []*Quiz
	if err := r.db.
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&quizzes).Error; err != nil {
		return nil, err
	}
	return quizzes, nil
}
