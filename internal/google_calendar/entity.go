package googlecalendar

import (
	"time"

	"github.com/google/uuid"
)

type CalendarTask struct {
	ID                    uuid.UUID
	Name                  string
	Description           string
	StartDate             *time.Time
	DueDate               *time.Time
	GoogleCalendarEventID *string
}
