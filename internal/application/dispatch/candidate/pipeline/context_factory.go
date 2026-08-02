package pipeline

import (
	"time"

	"github.com/google/uuid"
)

func NewContext(
	rideID uuid.UUID,
	pickupLat float64,
	pickupLng float64,
	retryAttempt int,
) *Context {

	return &Context{
		RideID: rideID,

		PickupLat: pickupLat,
		PickupLng: pickupLng,

		RetryAttempt: retryAttempt,

		StartedAt: time.Now(),
	}
}
