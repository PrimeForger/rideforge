package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

type DriverPushTokenRepository struct {
	db *sql.DB
}

func NewDriverPushTokenRepository(db *sql.DB) *DriverPushTokenRepository {
	return &DriverPushTokenRepository{db: db}
}

func (r *DriverPushTokenRepository) UpsertTx(
	ctx context.Context,
	tx *sql.Tx,
	driverID uuid.UUID,
	deviceID string,
	platform string,
	token string,
) error {
	_, err := tx.ExecContext(ctx, driverPushTokenUpsertTx, driverID, deviceID, platform, token)

	return err
}

func (r *DriverPushTokenRepository) GetActiveTokens(
	ctx context.Context,
	driverID uuid.UUID,
) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, driverPushTokenGetActiveTokens, driverID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []string

	for rows.Next() {
		var token string
		if err := rows.Scan(&token); err != nil {
			return nil, err
		}
		tokens = append(tokens, token)
	}

	return tokens, rows.Err()
}
