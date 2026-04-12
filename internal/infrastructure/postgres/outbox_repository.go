package postgres

import (
	"context"
	"database/sql"

	"github.com/ashadashraf/ride-hail-app/internal/domain/outbox"
	"github.com/google/uuid"
)

type OutboxRepository struct {
	db *sql.DB
}

func NewOutboxRepository(db *sql.DB) *OutboxRepository {
	return &OutboxRepository{db: db}
}

func (o *OutboxRepository) Insert(ctx context.Context, tx *sql.Tx, e *outbox.Event) error {
	query := `
	INSERT INTO outbox_events (id, aggregate_id, event_type, payload, created_at, published, processed_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := tx.ExecContext(ctx, query,
		e.ID,
		e.AggregateID,
		e.EventType,
		e.Payload,
		e.CreatedAt,
		e.Published,
		e.ProcessedAt,
	)

	return err
}

func (o *OutboxRepository) GetUnpublishedTx(ctx context.Context, tx *sql.Tx) ([]*outbox.Event, error) {
	query := `
	SELECT id, aggregate_id, event_type, payload, created_at, published, processed_at
	FROM outbox_events
	WHERE published = false
	ORDER BY created_at ASC
	FOR UPDATE SKIP LOCKED
	LIMIT 50
	`

	rows, err := tx.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*outbox.Event

	for rows.Next() {
		var e outbox.Event

		if err := rows.Scan(
			&e.ID,
			&e.AggregateID,
			&e.EventType,
			&e.Payload,
			&e.CreatedAt,
			&e.Published,
			&e.ProcessedAt,
		); err != nil {
			return nil, err
		}

		events = append(events, &e)
	}

	return events, nil
}

func (o *OutboxRepository) MarkPublishedTx(ctx context.Context, tx *sql.Tx, id uuid.UUID) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE outbox_events 
		SET published = true, processed_at = NOW()
		WHERE id = $1`,
		id,
	)
	return err
}
