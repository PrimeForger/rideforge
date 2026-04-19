package ports

import (
	"context"

	"github.com/google/uuid"
)

type DriverLocker interface {
	Reserve(ctx context.Context, driverID uuid.UUID, rideID uuid.UUID) (bool, error)
	Release(ctx context.Context, driverID uuid.UUID) error
}
