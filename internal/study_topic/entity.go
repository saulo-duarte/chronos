package studytopic

import (
	"time"

	"github.com/google/uuid"
	studysubject "github.com/saulo-duarte/chronos-lambda/internal/study_subject"
	"github.com/saulo-duarte/chronos-lambda/internal/user"
)

type StudyTopic struct {
	ID             uuid.UUID                 `gorm:"type:uuid;default:uuid_generate_v4()" json:"id"`
	Name           string                    `json:"name"`
	Description    string                    `json:"description"`
	Position       int                       `gorm:"default:0" json:"position"`
	UserID         uuid.UUID                 `gorm:"column:user_id;not null" json:"user_id"`
	User           user.User                 `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"-"`
	StudySubjectID uuid.UUID                 `gorm:"column:subject_id;not null" json:"subject_id"`
	StudySubject   studysubject.StudySubject `gorm:"foreignKey:StudySubjectID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"-"`
	CreatedAt      time.Time                 `json:"created_at"`
	UpdatedAt      time.Time                 `json:"updated_at"`
}
