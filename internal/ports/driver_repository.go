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
	GetEligibleDriversTx(ctx context.Context, tx *sql.Tx, rideID uuid.UUID, driverIDs []uuid.UUID) ([]*driver.Driver, error)
	GetByID(ctx context.Context, id uuid.UUID) (*driver.Driver, error)
	GetByIDTx(ctx context.Context, tx *sql.Tx, id uuid.UUID) (*driver.Driver, error)
	GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*driver.Driver, error)
	Save(ctx context.Context, d *driver.Driver) error
	SaveTx(ctx context.Context, tx *sql.Tx, d *driver.Driver) error
	InsertRideOfferTx(ctx context.Context, tx *sql.Tx, rideID, driverID uuid.UUID, attempt int) error
	MarkDriverRejectedTx(ctx context.Context, tx *sql.Tx, rideID, driverID uuid.UUID) error
	MarkDriverTimeoutTx(ctx context.Context, tx *sql.Tx, rideID, driverID uuid.UUID) error
	CountRideAttemptsTx(ctx context.Context, tx *sql.Tx, rideID uuid.UUID) (int, error)
}
