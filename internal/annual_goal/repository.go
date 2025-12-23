package annual_goal

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	Create(goal *AnnualGoal) error
	FindAllByUserID(userID uuid.UUID) ([]AnnualGoal, error)
	FindByID(id uuid.UUID) (*AnnualGoal, error)
	Update(goal *AnnualGoal) error
	Delete(id uuid.UUID) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(goal *AnnualGoal) error {
	return r.db.Create(goal).Error
}

func (r *repository) FindAllByUserID(userID uuid.UUID) ([]AnnualGoal, error) {
	var goals []AnnualGoal
	if err := r.db.Where("user_id = ?", userID).Find(&goals).Error; err != nil {
		return nil, err
	}
	return goals, nil
}

func (r *repository) FindByID(id uuid.UUID) (*AnnualGoal, error) {
	var goal AnnualGoal
	if err := r.db.First(&goal, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &goal, nil
}

func (r *repository) Update(goal *AnnualGoal) error {
	return r.db.Save(goal).Error
}

func (r *repository) Delete(id uuid.UUID) error {
	return r.db.Delete(&AnnualGoal{}, "id = ?", id).Error
}
