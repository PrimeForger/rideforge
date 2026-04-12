package application

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type IdempotencyService struct {
	db *sql.DB
}

func NewIdempotencyService(db *sql.DB) *IdempotencyService {
	return &IdempotencyService{db: db}
}

func (s *IdempotencyService) IsProcessed(ctx context.Context, eventID string) (bool, error) {
	id, err := uuid.Parse(eventID)
	if err != nil {
		return false, err
	}

	_, err = s.db.ExecContext(ctx,
		`INSERT INTO processed_events (id, processed_at)
		 VALUES ($1, NOW())`,
		id,
	)

	if err != nil {
		// duplicate → already processed
		if isUniqueViolation(err) {
			return true, nil
		}
		return false, err
	}

	return false, nil
}

func isUniqueViolation(err error) bool {
	// postgres error code 23505
	if pqErr, ok := err.(*pq.Error); ok {
		return pqErr.Code == "23505"
	}
	return false
}
