package ports

import (
	"context"
	"database/sql"

	"github.com/ashadashraf/ride-hail-app/internal/domain/ride"
	"github.com/google/uuid"
)

type RideRepository interface {
	Save(ctx context.Context, r *ride.Ride) error
	SaveTx(ctx context.Context, tx *sql.Tx, rideEntity *ride.Ride) error
	GetByID(ctx context.Context, id uuid.UUID) (*ride.Ride, error)
	GetByIDTx(ctx context.Context, tx *sql.Tx, id uuid.UUID) (*ride.Ride, error)
}
