package ports

import (
	"context"

	"github.com/ashadashraf/ride-hail-app/internal/domain/ride"
	"github.com/google/uuid"
)

type RideRepository interface {
	Save(ctx context.Context, r *ride.Ride) error
	GetByID(ctx context.Context, id uuid.UUID) (*ride.Ride, error)
}
