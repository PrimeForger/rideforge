package ports

import (
	"context"
	"database/sql"

	"github.com/ashadashraf/ride-hail-app/internal/domain/driver"
	"github.com/google/uuid"
)

type DriverRepository interface {
	GetAvailableDrivers(ctx context.Context) ([]*driver.Driver, error)
	GetAvailableDriversExcludingTx(ctx context.Context, tx *sql.Tx, rideID uuid.UUID) ([]*driver.Driver, error)
	GetByID(ctx context.Context, id uuid.UUID) (*driver.Driver, error)
	Save(ctx context.Context, d *driver.Driver) error
	InsertRideOfferTx(ctx context.Context, tx *sql.Tx, rideID uuid.UUID, driverID uuid.UUID, attempt int) error
	MarkDriverRejectedTx(ctx context.Context, tx *sql.Tx, rideID uuid.UUID, driverID uuid.UUID) error
	CountRideAttemptsTx(ctx context.Context, tx *sql.Tx, rideID uuid.UUID) (int, error)
}
