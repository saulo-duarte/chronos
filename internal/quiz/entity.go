package quiz

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type Quiz struct {
	ID             uuid.UUID  `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	UserID         uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	SubjectID      uuid.UUID  `gorm:"type:uuid;not null;index" json:"subject_id"`
	Topic          string     `gorm:"type:text;not null" json:"topic"`
	TotalQuestions int        `gorm:"not null;default:0" json:"total_questions"`
	CorrectCount   int        `gorm:"not null;default:0" json:"correct_count"`
	CreatedAt      time.Time  `gorm:"autoCreateTime" json:"created_at"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`

	Questions []QuizQuestion `gorm:"foreignKey:QuizID;constraint:OnDelete:CASCADE" json:"questions,omitempty"`
}

type QuizQuestion struct {
	ID            uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	QuizID        uuid.UUID      `gorm:"type:uuid;not null;index" json:"quiz_id"`
	Content       string         `gorm:"type:text;not null" json:"content"`
	Options       datatypes.JSON `gorm:"type:jsonb;not null" json:"options"`
	CorrectAnswer string         `gorm:"type:text;not null" json:"correct_answer"`
	Explanation   *string        `gorm:"type:text" json:"explanation,omitempty"`
	OrderIndex    int            `gorm:"not null" json:"order_index"`
	CreatedAt     time.Time      `gorm:"autoCreateTime" json:"created_at"`
}
