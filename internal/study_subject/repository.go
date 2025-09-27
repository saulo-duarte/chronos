package studysubject

import (
	"errors"

	"gorm.io/gorm"
)

type StudySubjectRepository interface {
	Create(s *StudySubject) error
	ListByUser(userID string) ([]*StudySubject, error)
	Update(s *StudySubject) error
	Delete(id string) error
	GetByID(id string) (*StudySubject, error)
}

type studySubjectRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) StudySubjectRepository {
	return &studySubjectRepository{db: db}
}

func (r *studySubjectRepository) Create(s *StudySubject) error {
	return r.db.Create(s).Error
}

func (r *studySubjectRepository) ListByUser(userID string) ([]*StudySubject, error) {
	var subjects []*StudySubject
	if err := r.db.Where("user_id = ?", userID).Find(&subjects).Error; err != nil {
		return nil, err
	}
	return subjects, nil
}

func (r *studySubjectRepository) Update(s *StudySubject) error {
	return r.db.Save(s).Error
}

func (r *studySubjectRepository) Delete(id string) error {
	return r.db.Delete(&StudySubject{}, "id = ?", id).Error
}

func (r *studySubjectRepository) GetByID(id string) (*StudySubject, error) {
	var subject StudySubject
	if err := r.db.First(&subject, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &subject, nil
}
