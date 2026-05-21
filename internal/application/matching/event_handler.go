package matching

import (
	"context"

	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/redis"
	"github.com/google/uuid"
)

type RideEventHandler struct {
	timeoutScheduler *redis.TimeoutScheduler
}

func NewRideEventHandler(
	scheduler *redis.TimeoutScheduler,
) *RideEventHandler {
	return &RideEventHandler{
		timeoutScheduler: scheduler,
	}
}

func (h *RideEventHandler) HandleRideAccepted(
	ctx context.Context,
	rideID uuid.UUID,
) error {

	// Cancel ALL pending timeouts for this ride
	return h.timeoutScheduler.CancelAll(ctx, rideID)
}

func (h *RideEventHandler) HandleDriverRejected(
	ctx context.Context,
	rideID uuid.UUID,
	driverID uuid.UUID,
) error {

	// Cancel only this driver's timeout
	return h.timeoutScheduler.Cancel(ctx, rideID, driverID)
}
