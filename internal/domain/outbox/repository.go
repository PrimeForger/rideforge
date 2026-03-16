package outbox

import "context"

type Repository interface {
	Save(ctx context.Context, event Event) error
	GetUnprocessed(ctx context.Context) ([]Event, error)
	MarkProcessed(ctx context.Context, id string) error
}
