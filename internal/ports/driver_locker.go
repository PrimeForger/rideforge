package ports

import (
	"context"

	"github.com/google/uuid"
)

type DriverLocker interface {
	Reserve(ctx context.Context, driverID uuid.UUID, rideID uuid.UUID) (bool, error)
	// ReserveTx(ctx context.Context, tx *sql.Tx, driverID uuid.UUID, rideID uuid.UUID) (bool, error)
	Release(ctx context.Context, driverID uuid.UUID, rideID uuid.UUID) (bool, error)
	// ReleaseTx(ctx context.Context, tx *sql.Tx, driverID uuid.UUID, rideID uuid.UUID) (bool, error)
	ForceRelease(ctx context.Context, driverID uuid.UUID) error
}
