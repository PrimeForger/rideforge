package ports

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

type DriverPushTokenRepository interface {
	UpsertTx(
		ctx context.Context,
		tx *sql.Tx,
		driverID uuid.UUID,
		deviceID string,
		platform string,
		token string,
	) error

	GetActiveTokens(ctx context.Context, driverID uuid.UUID) ([]string, error)
}
