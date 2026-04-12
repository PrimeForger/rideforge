package ports

import (
	"context"
	"database/sql"

	"github.com/ashadashraf/ride-hail-app/internal/domain/outbox"
	"github.com/google/uuid"
)

type OutboxRepository interface {
	Insert(ctx context.Context, tx *sql.Tx, e *outbox.Event) error

	GetUnpublishedTx(ctx context.Context, tx *sql.Tx) ([]*outbox.Event, error)
	MarkPublishedTx(ctx context.Context, tx *sql.Tx, id uuid.UUID) error
}
