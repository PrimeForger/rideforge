package matching

import (
	"context"
	"time"

	"github.com/ashadashraf/ride-hail-app/internal/config"
	"github.com/ashadashraf/ride-hail-app/internal/domain/ride"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/redis"
	"github.com/ashadashraf/ride-hail-app/internal/ports"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
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

var driverOfferTracer = otel.Tracer("application.driver_offer")

func (h *DriverOfferHandler) HandleDriverOffered(
	ctx context.Context,
	rideID uuid.UUID,
	driverID uuid.UUID,
	offerTimeout time.Duration,
) error {
	if offerTimeout <= 0 {
		offerTimeout = h.cfg.DriverOfferTimeout
	}

	ctx, span := driverOfferTracer.Start(ctx, "DriverOfferHandler.HandleDriverOffered")
	defer span.End()

	span.SetAttributes(
		attribute.String("ride.id", rideID.String()),
		attribute.String("driver.id", driverID.String()),
		attribute.Int64("offer.timeout_ms", offerTimeout.Milliseconds()),
	)

	fail := func(err error, result string) error {
		span.RecordError(err)
		span.SetAttributes(attribute.String("driver_offer.result", result))
		span.SetStatus(codes.Error, result)
		return err
	}

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
		return fail(err, "mark_offer_delivery_failed")
	}

	if err := h.driverCache.MarkDriverOffered(ctx, rideID, driverID); err != nil {
		h.logger.Error("failed to mark driver offered",
			zap.String("ride_id", rideID.String()),
			zap.String("driver_id", driverID.String()),
			zap.Error(err),
		)
		return fail(err, "mark_driver_offered_failed")
	}

	if err := h.timeoutScheduler.Schedule(ctx, rideID, driverID, offerTimeout); err != nil {
		h.logger.Error("failed to schedule driver offer timeout",
			zap.String("ride_id", rideID.String()),
			zap.String("driver_id", driverID.String()),
			zap.Duration("timeout", offerTimeout),
			zap.Error(err),
		)
		return fail(err, "schedule_timeout_failed")
	}

	span.SetAttributes(attribute.Bool("offer.timeout_scheduled", true))

	h.logger.Info("driver offer timeout scheduled",
		zap.String("ride_id", rideID.String()),
		zap.String("driver_id", driverID.String()),
		zap.Duration("timeout", offerTimeout),
	)

	if err := h.offerGateway.SendOffer(ctx, rideID, driverID); err != nil {
		h.logger.Error("failed to send driver offer",
			zap.String("ride_id", rideID.String()),
			zap.String("driver_id", driverID.String()),
			zap.Error(err),
		)

		return fail(err, "send_offer_failed")
	}

	span.SetAttributes(attribute.String("driver_offer.result", "sent"))
	span.SetStatus(codes.Ok, "sent")

	h.logger.Info("driver offer sent",
		zap.String("ride_id", rideID.String()),
		zap.String("driver_id", driverID.String()),
	)

	return nil
}
