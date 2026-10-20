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

	res, err := l.db.ExecContext(ctx, driverLockerReserve, driverID, rideID)

	if err != nil {
		return false, err
	}

	rows, _ := res.RowsAffected()

	return rows == 1, nil
}

func (l *DBDriverLocker) ReserveTx(ctx context.Context, tx *sql.Tx, driverID uuid.UUID, rideID uuid.UUID) (bool, error) {

	res, err := tx.ExecContext(ctx, driverLockerReserve, driverID, rideID)

	if err != nil {
		return false, err
	}

	rows, _ := res.RowsAffected()

	return rows == 1, nil
}

func (l *DBDriverLocker) Release(ctx context.Context, driverID uuid.UUID, rideID uuid.UUID) (bool, error) {
	res, err := l.db.ExecContext(ctx, driverLockerRelease, driverID, rideID)

	if err != nil {
		return false, err
	}

	rows, _ := res.RowsAffected()

	return rows != 0, nil
}

func (l *DBDriverLocker) ReleaseTx(ctx context.Context, tx *sql.Tx, driverID uuid.UUID, rideID uuid.UUID) (bool, error) {
	res, err := tx.ExecContext(ctx, driverLockerRelease, driverID, rideID)

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

	_, err := l.db.ExecContext(ctx, driverLockerForceRelease, driverID)

	return err
}
