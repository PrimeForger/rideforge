package ports

import (
	"context"

	"github.com/ashadashraf/ride-hail-app/internal/application/events"
)

type EventBus interface {
	Publish(ctx context.Context, e events.Envelope) error
}
