package task

import (
	"context"
)

type EventHandler interface {
	HandleTaskEvent(ctx context.Context, t *Task, accessToken string) error
}
