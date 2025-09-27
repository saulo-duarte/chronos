package project

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProjectRepository interface {
	Create(p *Project) error
	GetByID(id string) (*Project, error)
	ListByUser(userID uuid.UUID) ([]*Project, error)
	Update(p *Project) error
	Delete(id string) error
}

type projectRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) ProjectRepository {
	return &projectRepository{db: db}
}

func (r *projectRepository) Create(p *Project) error {
	return r.db.Create(p).Error
}

func (r *projectRepository) GetByID(id string) (*Project, error) {
	var p Project
	if err := r.db.First(&p, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (r *projectRepository) ListByUser(userID uuid.UUID) ([]*Project, error) {
	var projects []*Project
	if err := r.db.Where("user_id = ?", userID).Find(&projects).Error; err != nil {
		return nil, err
	}
	return projects, nil
}

func (r *projectRepository) Update(p *Project) error {
	return r.db.Save(p).Error
}

func (r *projectRepository) Delete(id string) error {
	return r.db.Delete(&Project{}, "id = ?", id).Error
}
