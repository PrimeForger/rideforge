package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

type ProcessedEventRepository struct {
	db *sql.DB
}

func NewProcessedEventRepository(db *sql.DB) *ProcessedEventRepository {
	return &ProcessedEventRepository{db: db}
}

func (r *ProcessedEventRepository) InsertIfNotExists(
	ctx context.Context,
	tx *sql.Tx,
	eventID uuid.UUID,
	consumer string,
) (bool, error) {

	query := `
	INSERT INTO processed_events (event_id, consumer_name)
	VALUES ($1, $2)
	ON CONFLICT DO NOTHING
	`

	res, err := tx.ExecContext(ctx, query, eventID, consumer)
	if err != nil {
		return false, err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return false, err
	}

	// If 0 rows -> already processed
	return rowsAffected > 0, nil
}
