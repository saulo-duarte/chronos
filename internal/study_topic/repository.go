package studytopic

import (
	"errors"

	"gorm.io/gorm"
)

type StudyTopicRepository interface {
	Create(t *StudyTopic) error
	GetByID(id string) (*StudyTopic, error)
	ListBySubject(studySubjectID string) ([]*StudyTopic, error)
	Update(t *StudyTopic) error
	Delete(id string) error
}

type studyTopicRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) StudyTopicRepository {
	return &studyTopicRepository{db: db}
}

func (r *studyTopicRepository) Create(t *StudyTopic) error {
	return r.db.Create(t).Error
}

func (r *studyTopicRepository) GetByID(id string) (*StudyTopic, error) {
	var topic StudyTopic
	if err := r.db.First(&topic, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &topic, nil
}

func (r *studyTopicRepository) ListBySubject(studySubjectID string) ([]*StudyTopic, error) {
	var topics []*StudyTopic
	if err := r.db.Where("subject_id = ?", studySubjectID).Find(&topics).Error; err != nil {
		return nil, err
	}
	return topics, nil
}

func (r *studyTopicRepository) Update(t *StudyTopic) error {
	return r.db.Save(t).Error
}

func (r *studyTopicRepository) Delete(id string) error {
	return r.db.Delete(&StudyTopic{}, "id = ?", id).Error
}
