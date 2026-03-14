package ports

import (
	"context"

	"github.com/ashadashraf/ride-hail-app/internal/domain/events"
)

type EventBus interface {
	Publish(ctx context.Context, e events.Event) error
}
