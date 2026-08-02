package stage

import (
	"context"

	"github.com/google/uuid"
)

type OfferedDriverProvider interface {
	GetOfferedDrivers(
		ctx context.Context,
		rideID uuid.UUID,
	) (map[uuid.UUID]struct{}, error)
}
