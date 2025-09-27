package task

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrNotFound = errors.New("task not found")
)

type TaskRepository interface {
	Create(t *Task) error
	FindByIdAndUserId(id, userId uuid.UUID) (*Task, error)
	ListByUser(userId uuid.UUID) ([]*Task, error)
	ListByProjectAndUser(projectId, userId uuid.UUID) ([]*Task, error)
	ListByStudyTopicAndUser(topicId, userId uuid.UUID) ([]*Task, error)
	Update(t *Task) error
	Delete(id, userId uuid.UUID) error
}

type taskRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) TaskRepository {
	return &taskRepository{db: db}
}

func (r *taskRepository) Create(t *Task) error {
	return r.db.Create(t).Error
}

func (r *taskRepository) FindByIdAndUserId(id, userId uuid.UUID) (*Task, error) {
	var t Task
	if err := r.db.Where("id = ? AND user_id = ?", id, userId).First(&t).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &t, nil
}

func (r *taskRepository) ListByUser(userId uuid.UUID) ([]*Task, error) {
	var tasks []*Task
	if err := r.db.Preload("Project").Preload("StudyTopic").Where("user_id = ?", userId).Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

func (r *taskRepository) ListByProjectAndUser(projectId, userId uuid.UUID) ([]*Task, error) {
	var tasks []*Task
	if err := r.db.Preload("Project").Preload("StudyTopic").Where("project_id = ? AND user_id = ?", projectId, userId).Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

func (r *taskRepository) ListByStudyTopicAndUser(topicId, userId uuid.UUID) ([]*Task, error) {
	var tasks []*Task
	if err := r.db.Preload("Project").Preload("StudyTopic").Where("study_topic_id = ? AND user_id = ?", topicId, userId).Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

func (r *taskRepository) Update(t *Task) error {
	return r.db.Save(t).Error
}

func (r *taskRepository) Delete(id, userId uuid.UUID) error {
	result := r.db.Where("id = ? AND user_id = ?", id, userId).Delete(&Task{})
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return result.Error
}
