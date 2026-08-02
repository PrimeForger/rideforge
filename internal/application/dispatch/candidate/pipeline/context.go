package pipeline

import (
	"time"

	"github.com/google/uuid"
)

type Context struct {
	RideID uuid.UUID

	PickupLat float64
	PickupLng float64

	RetryAttempt int

	StartedAt time.Time

	Result Result
}
