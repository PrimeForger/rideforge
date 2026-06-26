package matching

import (
	"context"

	"github.com/ashadashraf/ride-hail-app/internal/config"
	"github.com/ashadashraf/ride-hail-app/internal/domain/ride"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/redis"
	"github.com/ashadashraf/ride-hail-app/internal/ports"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type DriverOfferHandler struct {
	driverCache      *redis.DriverCache
	timeoutScheduler *redis.TimeoutScheduler
	offerGateway     ports.DriverOfferGateway
	cfg              *config.Config
	logger           *zap.Logger
}

func NewDriverOfferHandler(
	driverCache *redis.DriverCache,
	timeoutScheduler *redis.TimeoutScheduler,
	offerGateway ports.DriverOfferGateway,
	cfg *config.Config,
	logger *zap.Logger,
) *DriverOfferHandler {
	return &DriverOfferHandler{
		driverCache:      driverCache,
		timeoutScheduler: timeoutScheduler,
		offerGateway:     offerGateway,
		cfg:              cfg,
		logger:           logger,
	}
}

func (h *DriverOfferHandler) HandleDriverOffered(
	ctx context.Context,
	rideID uuid.UUID,
	driverID uuid.UUID,
) error {
	h.logger.Info("handling driver offered event",
		zap.String("ride_id", rideID.String()),
		zap.String("driver_id", driverID.String()),
	)

	if err := h.driverCache.MarkOfferDeliveryStatus(ctx, rideID, driverID, ride.OfferDeliveryOffered); err != nil {
		h.logger.Error("failed to mark offer delivery status",
			zap.String("ride_id", rideID.String()),
			zap.String("driver_id", driverID.String()),
			zap.Error(err),
		)
		return err
	}

	if err := h.driverCache.MarkDriverOffered(ctx, rideID, driverID); err != nil {
		h.logger.Error("failed to mark driver offered",
			zap.String("ride_id", rideID.String()),
			zap.String("driver_id", driverID.String()),
			zap.Error(err),
		)
		return err
	}

	if err := h.timeoutScheduler.Schedule(ctx, rideID, driverID, h.cfg.DriverOfferTimeout); err != nil {
		h.logger.Error("failed to schedule driver offer timeout",
			zap.String("ride_id", rideID.String()),
			zap.String("driver_id", driverID.String()),
			zap.Duration("timeout", h.cfg.DriverOfferTimeout),
			zap.Error(err),
		)
		return err
	}

	h.logger.Info("driver offer timeout scheduled",
		zap.String("ride_id", rideID.String()),
		zap.String("driver_id", driverID.String()),
		zap.Duration("timeout", h.cfg.DriverOfferTimeout),
	)

	err := h.offerGateway.SendOffer(ctx, rideID, driverID)
	if err != nil {
		h.logger.Error("failed to send driver offer",
			zap.String("ride_id", rideID.String()),
			zap.String("driver_id", driverID.String()),
			zap.Error(err),
		)
		return err
	}

	h.logger.Info("driver offer sent",
		zap.String("ride_id", rideID.String()),
		zap.String("driver_id", driverID.String()),
	)

	return nil
}
