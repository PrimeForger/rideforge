package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

type DBDriverLocker struct {
	db *sql.DB
}

func NewDBDriverLocker(db *sql.DB) *DBDriverLocker {
	return &DBDriverLocker{db: db}
}

func (l *DBDriverLocker) Reserve(ctx context.Context, driverID uuid.UUID, rideID uuid.UUID) (bool, error) {

	res, err := l.db.ExecContext(ctx, `
		UPDATE drivers
		SET status = 'RESERVED',
			reserved_for_ride = $2,
			reserved_at = NOW()
		WHERE id = $1 AND status = 'ONLINE'
	`, driverID, rideID)

	if err != nil {
		return false, err
	}

	rows, _ := res.RowsAffected()

	return rows == 1, nil
}

func (l *DBDriverLocker) ReserveTx(ctx context.Context, tx *sql.Tx, driverID uuid.UUID, rideID uuid.UUID) (bool, error) {

	res, err := tx.ExecContext(ctx, `
		UPDATE drivers
		SET status = 'RESERVED',
			reserved_for_ride = $2,
			reserved_at = NOW()
		WHERE id = $1 AND status = 'ONLINE'
	`, driverID, rideID)

	if err != nil {
		return false, err
	}

	rows, _ := res.RowsAffected()

	return rows == 1, nil
}

func (l *DBDriverLocker) Release(ctx context.Context, driverID uuid.UUID, rideID uuid.UUID) (bool, error) {
	res, err := l.db.ExecContext(ctx, `
		UPDATE drivers 
		SET status = 'ONLINE' 
			reserved_for_ride = NULL,
			reserved_at = NULL
		WHERE id = $1
		AND reserved_for_ride = $2
		AND status = 'RESERVED'
	`, driverID, rideID)

	if err != nil {
		return false, err
	}

	rows, _ := res.RowsAffected()

	return rows != 0, nil
}

func (l *DBDriverLocker) ReleaseTx(ctx context.Context, tx *sql.Tx, driverID uuid.UUID, rideID uuid.UUID) (bool, error) {
	res, err := tx.ExecContext(ctx, `
		UPDATE drivers 
		SET status = 'ONLINE' 
			reserved_for_ride = NULL,
			reserved_at = NULL
		WHERE id = $1
		AND reserved_for_ride = $2
		AND status = 'RESERVED'
	`, driverID, rideID)

	if err != nil {
		return false, err
	}

	rows, _ := res.RowsAffected()

	return rows != 0, nil
}

func (l *DBDriverLocker) ForceRelease(
	ctx context.Context,
	driverID uuid.UUID,
) error {

	_, err := l.db.ExecContext(ctx, `
		UPDATE drivers
		SET status = 'ONLINE',
			reserved_for_ride = NULL,
			reserved_at = NULL
		WHERE id = $1
		AND status = 'RESERVED'
	`, driverID)

	return err
}
