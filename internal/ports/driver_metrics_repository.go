package ports

import (
	"context"
	"database/sql"

	"github.com/ashadashraf/ride-hail-app/internal/domain/driver"
	"github.com/google/uuid"
)

type DriverMetricsRepository interface {
	InsertMetricEventTx(ctx context.Context, tx *sql.Tx, eventID uuid.UUID, driverID uuid.UUID, eventType string) (bool, error)

	IncrementOfferedTx(ctx context.Context, tx *sql.Tx, driverID uuid.UUID) error
	IncrementAckedTx(ctx context.Context, tx *sql.Tx, driverID uuid.UUID) error
	IncrementAcceptedTx(ctx context.Context, tx *sql.Tx, driverID uuid.UUID) error
	IncrementRejectedTx(ctx context.Context, tx *sql.Tx, driverID uuid.UUID) error
	IncrementTimeoutTx(ctx context.Context, tx *sql.Tx, driverID uuid.UUID) error
	IncrementCompletedTx(ctx context.Context, tx *sql.Tx, driverID uuid.UUID) error
	IncrementCancelledTx(ctx context.Context, tx *sql.Tx, driverID uuid.UUID) error

	RecalculateRatesTx(ctx context.Context, tx *sql.Tx, driverID uuid.UUID) error
	GetMetricsTx(ctx context.Context, tx *sql.Tx, driverID uuid.UUID) (*driver.DriverMetricsSnapshot, error)
}
