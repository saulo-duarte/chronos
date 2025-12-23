package annual_goal

import (
	"time"

	"github.com/google/uuid"
)

type CreateAnnualGoalDTO struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	Year        int    `json:"year" binding:"required"`
}

type UpdateAnnualGoalDTO struct {
	Title       *string           `json:"title"`
	Description *string           `json:"description"`
	Year        *int              `json:"year"`
	Status      *AnnualGoalStatus `json:"status"`
}

type AnnualGoalResponse struct {
	ID          uuid.UUID        `json:"id"`
	Title       string           `json:"title"`
	Description string           `json:"description,omitempty"`
	Year        int              `json:"year"`
	Status      AnnualGoalStatus `json:"status"`
	UserID      uuid.UUID        `json:"user_id"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}
