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
	searchRadiusKm float64,
	candidateLimit int,
) *Context {

	return &Context{
		RideID: rideID,

		PickupLat: pickupLat,
		PickupLng: pickupLng,

		RetryAttempt: retryAttempt,

		SearchRadiusKm: searchRadiusKm,
		CandidateLimit: candidateLimit,

		StartedAt: time.Now(),
	}
}
