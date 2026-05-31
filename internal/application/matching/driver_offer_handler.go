package matching

import (
	"context"

	"github.com/ashadashraf/ride-hail-app/internal/config"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/redis"
	"github.com/ashadashraf/ride-hail-app/internal/ports"
	"github.com/google/uuid"
)

type DriverOfferHandler struct {
	driverCache      *redis.DriverCache
	timeoutScheduler *redis.TimeoutScheduler
	offerGateway     ports.DriverOfferGateway
	cfg              *config.Config
}

func NewDriverOfferHandler(
	driverCache *redis.DriverCache,
	timeoutScheduler *redis.TimeoutScheduler,
	offerGateway ports.DriverOfferGateway,
	cfg *config.Config,
) *DriverOfferHandler {
	return &DriverOfferHandler{
		driverCache:      driverCache,
		timeoutScheduler: timeoutScheduler,
		offerGateway:     offerGateway,
		cfg:              cfg,
	}
}

func (h *DriverOfferHandler) HandleDriverOffered(
	ctx context.Context,
	rideID uuid.UUID,
	driverID uuid.UUID,
) error {

	if err := h.driverCache.MarkDriverOffered(ctx, rideID, driverID); err != nil {
		return err
	}

	if err := h.timeoutScheduler.Schedule(
		ctx,
		rideID,
		driverID,
		h.cfg.DriverOfferTimeout,
	); err != nil {
		return err
	}

	return h.offerGateway.SendOffer(ctx, rideID, driverID)
}
