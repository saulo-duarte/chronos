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
	ErrInvalidEventDates     = errors.New("task must have valid start or due dates")
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

// --- Private Helper Methods ---

func (s *calendarService) getUserTokens(ctx context.Context, userID uuid.UUID) (*oauth2.Token, error) {
	log := config.WithContext(ctx)

	u, err := s.userRepo.GetByID(userID.String())
	if err != nil {
		log.WithError(err).Error("Failed to retrieve user for calendar client")
		return nil, err
	}
	if u == nil {
		return nil, ErrUserNotFound
	}

	if u.EncryptedGoogleAccessToken == "" {
		return nil, ErrMissingCalendarTokens
	}

	accessToken, err := config.Decrypt(u.EncryptedGoogleAccessToken)
	if err != nil {
		log.WithError(err).Error("Failed to decrypt access token")
		return nil, ErrDecryptionFailed
	}

	refreshToken := ""
	if u.EncryptedGoogleRefreshToken != "" {
		refreshToken, err = config.Decrypt(u.EncryptedGoogleRefreshToken)
		if err != nil {
			log.WithError(err).Error("Failed to decrypt refresh token")
			return nil, ErrDecryptionFailed
		}
	}

	return &oauth2.Token{
		AccessToken:  accessToken,
		TokenType:    "Bearer",
		RefreshToken: refreshToken,
		Expiry:       time.Now().Add(-time.Hour), // Force refresh
	}, nil
}

func (s *calendarService) refreshTokenIfNeeded(ctx context.Context, token *oauth2.Token) (*oauth2.Token, error) {
	log := config.WithContext(ctx)

	tokenSource := s.oauthConfig.TokenSource(ctx, token)
	newToken, err := tokenSource.Token()
	if err != nil {
		log.WithError(err).Error("Failed to refresh Google token")
		return nil, err
	}

	if newToken.AccessToken != token.AccessToken {
		log.Info("Google token refreshed successfully")
		// TODO: Considere persistir o novo token aqui
	}

	return newToken, nil
}

func (s *calendarService) getCalendarClient(ctx context.Context, userID uuid.UUID) (*gcal.Service, error) {
	log := config.WithContext(ctx)

	token, err := s.getUserTokens(ctx, userID)
	if err != nil {
		return nil, err
	}

	token, err = s.refreshTokenIfNeeded(ctx, token)
	if err != nil {
		return nil, err
	}

	client := oauth2.NewClient(ctx, s.oauthConfig.TokenSource(ctx, token))

	srv, err := gcal.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.WithError(err).Error("Failed to create Calendar service client")
		return nil, err
	}

	return srv, nil
}

func (s *calendarService) buildCalendarEvent(task *CalendarTask) (*gcal.Event, error) {
	if task.StartDate == nil && task.DueDate == nil {
		return nil, ErrInvalidEventDates
	}

	event := &gcal.Event{
		Summary:     task.Name,
		Description: task.Description,
		Reminders: &gcal.EventReminders{
			UseDefault: false,
		},
	}

	// Define o horário de início
	if task.StartDate != nil {
		event.Start = &gcal.EventDateTime{
			DateTime: task.StartDate.Format(time.RFC3339),
		}
	} else if task.DueDate != nil {
		// Se não tem StartDate, usa 1 hora antes da DueDate
		event.Start = &gcal.EventDateTime{
			DateTime: task.DueDate.Add(-time.Hour).Format(time.RFC3339),
		}
	}

	// Define o horário de término
	if task.DueDate != nil {
		event.End = &gcal.EventDateTime{
			DateTime: task.DueDate.Format(time.RFC3339),
		}
	} else if task.StartDate != nil {
		// Se não tem DueDate, usa 1 hora depois da StartDate
		event.End = &gcal.EventDateTime{
			DateTime: task.StartDate.Add(time.Hour).Format(time.RFC3339),
		}
	}

	return event, nil
}

func (s *calendarService) isEventNotFoundError(err error) bool {
	if apiErr, ok := err.(*googleapi.Error); ok {
		return apiErr.Code == 404
	}
	return false
}

// --- Public Methods ---

func (s *calendarService) AddEventToCalendar(ctx context.Context, userID uuid.UUID, task *CalendarTask) (string, error) {
	log := config.WithContext(ctx)

	srv, err := s.getCalendarClient(ctx, userID)
	if err != nil {
		return "", err
	}

	event, err := s.buildCalendarEvent(task)
	if err != nil {
		if errors.Is(err, ErrInvalidEventDates) {
			log.Warnf("Task %s has no valid dates to create a calendar event", task.ID)
			return "", nil
		}
		return "", err
	}

	calEvent, err := srv.Events.Insert("primary", event).Context(ctx).Do()
	if err != nil {
		log.WithError(err).Error("Failed to insert calendar event")
		return "", err
	}

	log.WithField("event_id", calEvent.Id).Infof("Created calendar event for task %s", task.ID)
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

	event, err := s.buildCalendarEvent(task)
	if err != nil {
		if errors.Is(err, ErrInvalidEventDates) {
			log.Warnf("Task %s no longer has valid dates, should be deleted", task.ID)
			return s.DeleteEventFromCalendar(ctx, userID, *task.GoogleCalendarEventID)
		}
		return err
	}

	_, err = srv.Events.Update("primary", *task.GoogleCalendarEventID, event).Context(ctx).Do()
	if err != nil {
		if s.isEventNotFoundError(err) {
			log.Warnf("Calendar event %s not found, considering as already deleted", *task.GoogleCalendarEventID)
			return nil
		}
		log.WithError(err).Error("Failed to update calendar event")
		return err
	}

	log.WithField("event_id", *task.GoogleCalendarEventID).Infof("Updated calendar event for task %s", task.ID)
	return nil
}

func (s *calendarService) DeleteEventFromCalendar(ctx context.Context, userID uuid.UUID, googleEventID string) error {
	log := config.WithContext(ctx)

	if googleEventID == "" {
		return nil
	}

	srv, err := s.getCalendarClient(ctx, userID)
	if err != nil {
		return err
	}

	err = srv.Events.Delete("primary", googleEventID).Context(ctx).Do()
	if err != nil {
		if s.isEventNotFoundError(err) {
			log.Warnf("Calendar event %s not found, considering as already deleted", googleEventID)
			return nil
		}
		log.WithError(err).Error("Failed to delete calendar event")
		return err
	}

	log.WithField("event_id", googleEventID).Info("Deleted calendar event successfully")
	return nil
}
