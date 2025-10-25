package googlecalendar

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/saulo-duarte/chronos-lambda/internal/config"
)

var ErrNoValidDates = errors.New("task has no valid dates for calendar event")

type CalendarManager interface {
	SyncTask(ctx context.Context, userID uuid.UUID, task *CalendarTask) (eventID string, err error)
	RemoveTask(ctx context.Context, userID uuid.UUID, eventID string) error
}

type calendarManager struct {
	calendarService CalendarService
}

func NewCalendarManager(calendarService CalendarService) CalendarManager {
	return &calendarManager{
		calendarService: calendarService,
	}
}

func (m *calendarManager) SyncTask(ctx context.Context, userID uuid.UUID, task *CalendarTask) (string, error) {
	log := config.WithContext(ctx)

	hasValidDates := task.StartDate != nil || task.DueDate != nil
	hasEventID := task.GoogleCalendarEventID != nil && *task.GoogleCalendarEventID != ""

	if hasEventID && !hasValidDates {
		log.Infof("Task %s no longer has valid dates, deleting calendar event", task.ID)
		if err := m.calendarService.DeleteEventFromCalendar(ctx, userID, *task.GoogleCalendarEventID); err != nil {
			log.WithError(err).Warnf("Failed to delete calendar event for task %s", task.ID)
		}
		return "", nil
	}

	if !hasValidDates {
		return "", nil
	}

	if hasEventID {
		if err := m.calendarService.UpdateEventInCalendar(ctx, userID, task); err != nil {
			log.WithError(err).Warnf("Failed to update calendar event for task %s", task.ID)
			return *task.GoogleCalendarEventID, err
		}
		return *task.GoogleCalendarEventID, nil
	}

	eventID, err := m.calendarService.AddEventToCalendar(ctx, userID, task)
	if err != nil {
		log.WithError(err).Warnf("Failed to create calendar event for task %s", task.ID)
		return "", err
	}

	if eventID == "" {
		log.Warnf("Calendar service returned empty event ID for task %s", task.ID)
		return "", nil
	}

	log.Infof("Created calendar event %s for task %s", eventID, task.ID)
	return eventID, nil
}

func (m *calendarManager) RemoveTask(ctx context.Context, userID uuid.UUID, eventID string) error {
	if eventID == "" {
		return nil
	}

	log := config.WithContext(ctx)

	if err := m.calendarService.DeleteEventFromCalendar(ctx, userID, eventID); err != nil {
		log.WithError(err).Warnf("Failed to delete calendar event %s", eventID)
		return err
	}

	return nil
}
