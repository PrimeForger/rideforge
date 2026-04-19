package application

import (
	"context"

	"github.com/google/uuid"
)

type GeoService interface {
	NearbyDrivers(ctx context.Context, lat, lng float64, radius int) ([]uuid.UUID, error)
}
