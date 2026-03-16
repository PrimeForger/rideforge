package postgres

import (
	"context"
	"database/sql"

	"github.com/ashadashraf/ride-hail-app/internal/domain/outbox"
)

type OutboxRepository struct {
	db *sql.DB
}

func NewOutboxRepository(db *sql.DB) *OutboxRepository {
	return &OutboxRepository{db: db}
}

func (r *OutboxRepository) Save(ctx context.Context, e outbox.Event) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO outbox_events
		(id, aggregate_id, event_type, payload)
		VALUES ($1,$2,$3,$4)
	`,
		e.ID,
		e.AggregateID,
		e.EventType,
		e.Payload,
	)

	return err
}
