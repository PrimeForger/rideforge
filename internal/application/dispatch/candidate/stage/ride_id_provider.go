package stage

import (
	"context"

	"github.com/google/uuid"
)

type RideIDProvider interface {
	RideID(
		ctx context.Context,
	) uuid.UUID
}
