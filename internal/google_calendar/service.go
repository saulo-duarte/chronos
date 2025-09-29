package googlecalendar

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/saulo-duarte/chronos-lambda/internal/config"
	"github.com/saulo-duarte/chronos-lambda/internal/user"
	"golang.org/x/oauth2"
	gcal "google.golang.org/api/calendar/v3"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

var (
	ErrUserNotFound          = errors.New("user not found for calendar integration")
	ErrDecryptionFailed      = errors.New("failed to decrypt user's google token")
	ErrMissingCalendarTokens = errors.New("user has no google access token")
)

type CalendarService interface {
	AddEventToCalendar(ctx context.Context, userID uuid.UUID, task *CalendarTask) (string, error)
	UpdateEventInCalendar(ctx context.Context, userID uuid.UUID, task *CalendarTask) error
	DeleteEventFromCalendar(ctx context.Context, userID uuid.UUID, googleEventID string) error
}

type calendarService struct {
	userRepo    user.UserRepository
	oauthConfig *oauth2.Config
}

func NewCalendarService(userRepo user.UserRepository, oauthConfig *oauth2.Config) CalendarService {
	return &calendarService{
		userRepo:    userRepo,
		oauthConfig: oauthConfig,
	}
}

func (s *calendarService) getCalendarClient(ctx context.Context, userID uuid.UUID) (*gcal.Service, error) {
	log := config.WithContext(ctx)

	u, err := s.userRepo.GetByID(userID.String())
	if err != nil {
		log.WithError(err).Error("Failed to retrieve user for calendar client")
		return nil, err
	}
	if u == nil {
		return nil, ErrUserNotFound
	}

	accessTokenEncrypted := u.EncryptedGoogleAccessToken
	if accessTokenEncrypted == "" {
		return nil, ErrMissingCalendarTokens
	}

	accessToken, err := config.Decrypt(accessTokenEncrypted)
	if err != nil {
		log.WithError(err).Error("Failed to decrypt access token")
		return nil, ErrDecryptionFailed
	}

	token := &oauth2.Token{
		AccessToken:  accessToken,
		TokenType:    "Bearer",
		RefreshToken: u.EncryptedGoogleRefreshToken,
		Expiry:       time.Now().Add(-time.Hour),
	}

	tokenSource := s.oauthConfig.TokenSource(ctx, token)
	newToken, err := tokenSource.Token()
	if err != nil {
		log.WithError(err).Error("Failed to refresh Google token")
		return nil, err
	}

	if newToken.AccessToken != accessToken {
		log.Info("Google token refreshed and should be persisted")
	}

	client := oauth2.NewClient(ctx, tokenSource)
	srv, err := gcal.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.WithError(err).Error("Failed to create Calendar service client")
		return nil, err
	}

	return srv, nil
}

func (s *calendarService) buildCalendarEvent(task *CalendarTask) *gcal.Event {
	event := &gcal.Event{
		Summary:     task.Name,
		Description: task.Description,
		Reminders: &gcal.EventReminders{
			UseDefault: false,
		},
	}

	if task.DueDate != nil {
		event.End = &gcal.EventDateTime{
			DateTime: task.DueDate.Format(time.RFC3339),
		}
		if task.StartDate == nil {
			event.Start = &gcal.EventDateTime{
				DateTime: task.DueDate.Add(-time.Hour).Format(time.RFC3339),
			}
		}
	}

	if task.StartDate != nil {
		event.Start = &gcal.EventDateTime{
			DateTime: task.StartDate.Format(time.RFC3339),
		}
	}

	if event.Start == nil || event.End == nil {
		return nil
	}

	return event
}

func (s *calendarService) AddEventToCalendar(ctx context.Context, userID uuid.UUID, task *CalendarTask) (string, error) {
	log := config.WithContext(ctx)
	srv, err := s.getCalendarClient(ctx, userID)
	if err != nil {
		return "", err
	}

	event := s.buildCalendarEvent(task)
	if event == nil {
		log.Warnf("Task %s has no valid dates to create a calendar event", task.ID)
		return "", nil
	}

	calEvent, err := srv.Events.Insert("primary", event).Context(ctx).Do()
	if err != nil {
		log.WithError(err).Error("Failed to insert calendar event")
		return "", err
	}

	return calEvent.Id, nil
}

func (s *calendarService) UpdateEventInCalendar(ctx context.Context, userID uuid.UUID, task *CalendarTask) error {
	log := config.WithContext(ctx)
	if task.GoogleCalendarEventID == nil || *task.GoogleCalendarEventID == "" {
		return errors.New("cannot update event: missing Google Calendar Event ID")
	}

	srv, err := s.getCalendarClient(ctx, userID)
	if err != nil {
		return err
	}

	event := s.buildCalendarEvent(task)
	if event == nil {
		log.Warnf("Task %s no longer has valid dates, attempting to delete calendar event", task.ID)
		return s.DeleteEventFromCalendar(ctx, userID, *task.GoogleCalendarEventID)
	}

	_, err = srv.Events.Update("primary", *task.GoogleCalendarEventID, event).Context(ctx).Do()
	if err != nil {
		log.WithError(err).Error("Failed to update calendar event")
		return err
	}

	return nil
}

func (s *calendarService) DeleteEventFromCalendar(ctx context.Context, userID uuid.UUID, googleEventID string) error {
	log := config.WithContext(ctx)
	srv, err := s.getCalendarClient(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrMissingCalendarTokens) || errors.Is(err, ErrDecryptionFailed) {
			log.Warnf("Skipping Google Calendar deletion for event %s due to missing/invalid token", googleEventID)
			return nil
		}
		return err
	}

	err = srv.Events.Delete("primary", googleEventID).Context(ctx).Do()
	if err != nil {
		if apiErr, ok := err.(*googleapi.Error); ok && apiErr.Code == 404 {
			log.Warnf("Calendar event %s not found on Google, considering deleted.", googleEventID)
			return nil
		}
		log.WithError(err).Error("Failed to delete calendar event")
		return err
	}

	return nil
}
