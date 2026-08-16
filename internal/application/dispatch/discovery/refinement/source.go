package refinement

import (
	"context"

	"github.com/google/uuid"
)

type GeoRefinementSource interface {
	NearestDrivers(
		ctx context.Context,
		lat float64,
		lng float64,
		radiusKm float64,
		driverIDs []uuid.UUID,
	) ([]uuid.UUID, error)
}
