package googlecalendar

import (
	"os"

	"github.com/saulo-duarte/chronos-lambda/internal/user"
	"golang.org/x/oauth2"
	gcal "google.golang.org/api/calendar/v3"
)

type GoogleCalendarContainer struct {
	CalendarService CalendarService
}

func NewGoogleCalendarContainer(
	userRepo user.UserRepository,
) *GoogleCalendarContainer {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	redirectURL := os.Getenv("GOOGLE_REDIRECT_URL")

	oauthConfig := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{gcal.CalendarEventsScope, gcal.CalendarScope},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.google.com/o/oauth2/auth",
			TokenURL: "https://oauth2.googleapis.com/token",
		},
	}

	calendarService := NewCalendarService(userRepo, oauthConfig)

	return &GoogleCalendarContainer{
		CalendarService: calendarService,
	}
}
