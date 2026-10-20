package postgres

import (
	"context"
	"database/sql"

	"github.com/ashadashraf/ride-hail-app/internal/domain/driver"
	"github.com/google/uuid"
)

type DriverMetricsRepository struct {
	db *sql.DB
}

func NewDriverMetricsRepository(db *sql.DB) *DriverMetricsRepository {
	return &DriverMetricsRepository{db: db}
}

func (r *DriverMetricsRepository) InsertMetricEventTx(
	ctx context.Context,
	tx *sql.Tx,
	eventID uuid.UUID,
	driverID uuid.UUID,
	eventType string,
) (bool, error) {
	res, err := tx.ExecContext(ctx, driverMetricsInsertMetricEventTx, eventID, driverID, eventType)
	if err != nil {
		return false, err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return false, err
	}

	return rows == 1, nil
}

func (r *DriverMetricsRepository) IncrementOfferedTx(ctx context.Context, tx *sql.Tx, driverID uuid.UUID) error {
	_, err := tx.ExecContext(ctx, driverMetricsIncrementOfferedTx, driverID)

	return err
}

func (r *DriverMetricsRepository) IncrementAckedTx(ctx context.Context, tx *sql.Tx, driverID uuid.UUID) error {
	_, err := tx.ExecContext(ctx, driverMetricsIncrementAckedTx, driverID)

	return err
}

func (r *DriverMetricsRepository) IncrementAcceptedTx(ctx context.Context, tx *sql.Tx, driverID uuid.UUID) error {
	_, err := tx.ExecContext(ctx, driverMetricsIncrementAcceptedTx, driverID)

	return err
}

func (r *DriverMetricsRepository) IncrementRejectedTx(ctx context.Context, tx *sql.Tx, driverID uuid.UUID) error {
	_, err := tx.ExecContext(ctx, driverMetricsIncrementRejectedTx, driverID)

	return err
}

func (r *DriverMetricsRepository) IncrementTimeoutTx(ctx context.Context, tx *sql.Tx, driverID uuid.UUID) error {
	_, err := tx.ExecContext(ctx, driverMetricsIncrementTimeoutTx, driverID)

	return err
}

func (r *DriverMetricsRepository) IncrementCompletedTx(ctx context.Context, tx *sql.Tx, driverID uuid.UUID) error {
	_, err := tx.ExecContext(ctx, driverMetricsIncrementCompletedTx, driverID)

	return err
}

func (r *DriverMetricsRepository) IncrementCancelledTx(ctx context.Context, tx *sql.Tx, driverID uuid.UUID) error {
	_, err := tx.ExecContext(ctx, driverMetricsIncrementCancelledTx, driverID)

	return err
}

func (r *DriverMetricsRepository) RecalculateRatesTx(ctx context.Context, tx *sql.Tx, driverID uuid.UUID) error {
	_, err := tx.ExecContext(ctx, driverMetricsRecalculateRatesTx, driverID)

	return err
}

func (r *DriverMetricsRepository) GetMetricsTx(
	ctx context.Context,
	tx *sql.Tx,
	driverID uuid.UUID,
) (*driver.DriverMetricsSnapshot, error) {
	row := tx.QueryRowContext(ctx, driverMetricsGetMetricsTx, driverID)

	var m driver.DriverMetricsSnapshot

	err := row.Scan(
		&m.DriverID,
		&m.AcceptanceRate,
		&m.CancellationRate,
		&m.TimeoutRate,
		&m.CompletedRides,
	)
	if err != nil {
		return nil, err
	}

	return &m, nil
}
