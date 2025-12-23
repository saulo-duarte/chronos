package annual_goal

import (
	"time"

	"github.com/google/uuid"
	"github.com/saulo-duarte/chronos-lambda/internal/user"
)

type AnnualGoal struct {
	ID          uuid.UUID        `gorm:"type:uuid;default:uuid_generate_v4()" json:"id"`
	Title       string           `json:"title"`
	Description string           `json:"description,omitempty"`
	Year        int              `json:"year"`
	Status      AnnualGoalStatus `json:"status"`
	UserID      uuid.UUID        `gorm:"column:user_id;not null" json:"user_id"`
	User        user.User        `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"-"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}
