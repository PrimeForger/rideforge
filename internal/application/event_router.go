package application

import (
	"context"
	"database/sql"
	"errors"
	"time"

	appevents "github.com/ashadashraf/ride-hail-app/internal/application/events"
	"github.com/ashadashraf/ride-hail-app/internal/application/matching"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/observability"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/postgres"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/redis"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

type EventRouter struct {
	txManager          *postgres.TxManager
	processedEventRepo *postgres.ProcessedEventRepository

	rideService           *RideService
	matchingEngine        *MatchingEngine
	rideEventHandler      *matching.RideEventHandler
	driverOfferHandler    *matching.DriverOfferHandler
	driverResponseService *DriverResponseService
	driverMetricsService  *DriverMetricsService

	geoService   *redis.GeoService
	driverCache  *redis.DriverCache
	h3Index      *redis.H3DriverIndex
	driverLocker interface {
		ForceRelease(ctx context.Context, driverID uuid.UUID) error
	}

	logger *zap.Logger
}

func NewEventRouter(
	txManager *postgres.TxManager,
	processedEventRepo *postgres.ProcessedEventRepository,
	rideService *RideService,
	matchingEngine *MatchingEngine,
	rideEventHandler *matching.RideEventHandler,
	driverOfferHandler *matching.DriverOfferHandler,
	driverResponseService *DriverResponseService,
	driverMetricsService *DriverMetricsService,
	geoService *redis.GeoService,
	driverCache *redis.DriverCache,
	h3Index *redis.H3DriverIndex,
	driverLocker interface {
		ForceRelease(ctx context.Context, driverID uuid.UUID) error
	},
	logger *zap.Logger,
) *EventRouter {
	return &EventRouter{
		txManager:             txManager,
		processedEventRepo:    processedEventRepo,
		rideService:           rideService,
		matchingEngine:        matchingEngine,
		rideEventHandler:      rideEventHandler,
		driverOfferHandler:    driverOfferHandler,
		driverResponseService: driverResponseService,
		driverMetricsService:  driverMetricsService,
		geoService:            geoService,
		driverCache:           driverCache,
		h3Index:               h3Index,
		driverLocker:          driverLocker,
		logger:                logger,
	}
}

var eventRouterTracer = otel.Tracer("application.event_router")

func (r *EventRouter) Handle(ctx context.Context, envelope appevents.Envelope) error {
	ctx, span := eventRouterTracer.Start(ctx, "EventRouter.Handle")
	defer span.End()

	span.SetAttributes(
		attribute.String("event.id", envelope.ID),
		attribute.String("event.type", envelope.Type),
		attribute.String("event.aggregate", envelope.Aggregate),
	)

	fail := func(err error, result string) error {
		span.RecordError(err)
		span.SetAttributes(attribute.String("event_router.result", result))
		span.SetStatus(codes.Error, result)
		return err
	}

	if envelope.ID == "" {
		return fail(errors.New("event id is required"), "missing_event_id")
	}

	if envelope.Type == "" {
		return fail(errors.New("event type is required"), "missing_event_type")
	}

	var err error

	switch envelope.Type {
	case "ride.requested", "matching.started", "matching.retry",
		"driver.accepted", "driver.rejected", "driver.timeout":

		span.SetAttributes(attribute.Bool("event.transactional", true))
		err = r.handleTransactional(ctx, envelope)

	default:
		span.SetAttributes(attribute.Bool("event.transactional", false))
		err = r.handleSideEffect(ctx, envelope)
	}

	if err != nil {
		return fail(err, "handler_error")
	}

	span.SetAttributes(attribute.String("event_router.result", "success"))
	span.SetStatus(codes.Ok, "success")
	return nil
}

func (r *EventRouter) handleTransactional(
	ctx context.Context,
	envelope appevents.Envelope,
) error {
	ctx, span := eventRouterTracer.Start(ctx, "EventRouter.handleTransactional")
	defer span.End()

	span.SetAttributes(
		attribute.String("event.id", envelope.ID),
		attribute.String("event.type", envelope.Type),
	)

	return r.txManager.WithinTx(ctx, func(tx *sql.Tx) error {

		eventID, err := uuid.Parse(envelope.ID)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "invalid_event_id")
			return err
		}

		processed, err := r.processedEventRepo.InsertIfNotExistsTx(
			ctx,
			tx,
			eventID,
			"matching-service",
		)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "idempotency_check_failed")
			return err
		}

		if !processed {
			span.SetAttributes(attribute.Bool("event.duplicate", true))
			span.SetStatus(codes.Ok, "duplicate skipped")

			r.logger.Info(
				"duplicate event skipped",
				zap.String("event_id", envelope.ID),
				zap.String("event_type", envelope.Type),
			)
			return nil
		}

		span.SetAttributes(attribute.Bool("event.duplicate", false))

		if err := r.dispatchTransactional(ctx, tx, envelope); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "transactional_dispatch_failed")
			return err
		}

		span.SetStatus(codes.Ok, "transactional processed")
		return nil
	})
}

func (r *EventRouter) dispatchTransactional(
	ctx context.Context,
	tx *sql.Tx,
	envelope appevents.Envelope,
) error {
	switch envelope.Type {
	case "ride.requested":
		var data struct {
			RideID string `json:"ride_id"`
		}
		if err := decodeEventData(envelope, &data); err != nil {
			return err
		}

		rideID, err := parseUUID(data.RideID, "ride_id")
		if err != nil {
			return err
		}

		return r.rideService.StartMatchingTx(ctx, tx, rideID)

	case "matching.started", "matching.retry":
		var data struct {
			RideID string `json:"ride_id"`
		}
		if err := decodeEventData(envelope, &data); err != nil {
			return err
		}

		rideID, err := parseUUID(data.RideID, "ride_id")
		if err != nil {
			return err
		}

		return r.matchingEngine.HandleMatchingStarted(ctx, tx, rideID)

	case "driver.accepted":
		var data driverRideEventData
		if err := decodeEventData(envelope, &data); err != nil {
			return err
		}

		rideID, driverID, err := parseRideDriverIDs(data)
		if err != nil {
			return err
		}

		if err := r.driverResponseService.HandleDriverAccepted(ctx, tx, rideID, driverID); err != nil {
			return err
		}

		observability.DriverResponsesTotal.WithLabelValues("accepted").Inc()

		return r.driverMetricsService.HandleDriverAccepted(ctx, envelope, driverID)

	case "driver.rejected":
		var data driverRideEventData
		if err := decodeEventData(envelope, &data); err != nil {
			return err
		}

		rideID, driverID, err := parseRideDriverIDs(data)
		if err != nil {
			return err
		}

		return r.driverResponseService.HandleDriverRejected(ctx, tx, rideID, driverID)

	case "driver.timeout":
		var data driverRideEventData
		if err := decodeEventData(envelope, &data); err != nil {
			return err
		}

		rideID, driverID, err := parseRideDriverIDs(data)
		if err != nil {
			return err
		}

		acked, err := r.driverCache.IsOfferAcked(ctx, rideID, driverID)
		if err != nil {
			return err
		}

		deliveryStatus, err := r.driverCache.GetOfferDeliveryStatus(ctx, rideID, driverID)
		if err != nil {
			return err
		}

		return r.driverResponseService.HandleDriverTimeout(
			ctx,
			tx,
			rideID,
			driverID,
			acked,
			string(deliveryStatus),
		)

	default:
		return nil
	}
}

func (r *EventRouter) handleSideEffect(
	ctx context.Context,
	envelope appevents.Envelope,
) error {
	ctx, span := eventRouterTracer.Start(ctx, "EventRouter.handleSideEffect")
	defer span.End()

	span.SetAttributes(
		attribute.String("event.id", envelope.ID),
		attribute.String("event.type", envelope.Type),
	)

	eventID, err := uuid.Parse(envelope.ID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid_event_id")
		return err
	}

	processed, err := r.processedEventRepo.InsertIfNotExists(
		ctx,
		eventID,
		"matching-service",
	)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "idempotency_check_failed")
		return err
	}

	if !processed {
		span.SetAttributes(attribute.Bool("event.duplicate", true))
		span.SetStatus(codes.Ok, "duplicate skipped")

		r.logger.Info(
			"duplicate event skipped",
			zap.String("event_id", envelope.ID),
			zap.String("event_type", envelope.Type),
		)
		return nil
	}

	span.SetAttributes(attribute.Bool("event.duplicate", false))

	err = r.dispatchSideEffect(ctx, envelope)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "side_effect_dispatch_failed")
		return err
	}

	span.SetStatus(codes.Ok, "side effect processed")
	return nil
}

func (r *EventRouter) dispatchSideEffect(
	ctx context.Context,
	envelope appevents.Envelope,
) error {
	switch envelope.Type {
	case "ride.accepted":
		var data struct {
			RideID string `json:"ride_id"`
		}
		if err := decodeEventData(envelope, &data); err != nil {
			return err
		}

		rideID, err := parseUUID(data.RideID, "ride_id")
		if err != nil {
			return err
		}

		return r.rideEventHandler.HandleRideAccepted(ctx, rideID)

	case "driver.online":
		var data struct {
			DriverID string  `json:"driver_id"`
			Lat      float64 `json:"lat"`
			Lng      float64 `json:"lng"`
		}
		if err := decodeEventData(envelope, &data); err != nil {
			return err
		}

		driverID, err := parseUUID(data.DriverID, "driver_id")
		if err != nil {
			return err
		}

		if err := r.geoService.UpdateDriverLocation(ctx, driverID, data.Lat, data.Lng); err != nil {
			return err
		}

		return r.driverCache.MarkOnline(ctx, driverID, data.Lat, data.Lng)

	case "driver.offline":
		var data struct {
			DriverID string `json:"driver_id"`
			Reason   string `json:"reason"`
		}
		if err := decodeEventData(envelope, &data); err != nil {
			return err
		}

		driverID, err := parseUUID(data.DriverID, "driver_id")
		if err != nil {
			return err
		}

		if err := r.geoService.RemoveDriver(ctx, driverID); err != nil {
			return err
		}

		if _, err := r.h3Index.RemoveDriver(ctx, driverID); err != nil {
			return err
		}

		if err := r.driverCache.MarkOffline(ctx, driverID); err != nil {
			return err
		}

		return r.driverLocker.ForceRelease(ctx, driverID)

	case "driver.offered":
		var data struct {
			RideID         string `json:"ride_id"`
			DriverID       string `json:"driver_id"`
			OfferTimeoutMs int64  `json:"offer_timeout_ms"`
		}

		if err := decodeEventData(envelope, &data); err != nil {
			return err
		}

		rideID, driverID, err := parseRideDriverIDs(driverRideEventData{
			RideID:   data.RideID,
			DriverID: data.DriverID,
		})
		if err != nil {
			return err
		}

		timeout := time.Duration(data.OfferTimeoutMs) * time.Millisecond

		if err := r.driverOfferHandler.HandleDriverOffered(ctx, rideID, driverID, timeout); err != nil {
			return err
		}

		return r.driverMetricsService.HandleDriverOffered(ctx, envelope, driverID)

	case "driver.offer.acked":
		var data struct {
			DriverID string `json:"driver_id"`
		}
		if err := decodeEventData(envelope, &data); err != nil {
			return err
		}

		driverID, err := parseUUID(data.DriverID, "driver_id")
		if err != nil {
			return err
		}

		return r.driverMetricsService.HandleDriverOfferAcked(ctx, envelope, driverID)

	case "driver.rejected.processed":
		var data driverRideEventData
		if err := decodeEventData(envelope, &data); err != nil {
			return err
		}

		rideID, driverID, err := parseRideDriverIDs(data)
		if err != nil {
			return err
		}

		if err := r.rideEventHandler.HandleDriverRejected(ctx, rideID, driverID); err != nil {
			return err
		}

		observability.DriverResponsesTotal.WithLabelValues("rejected").Inc()

		return r.driverMetricsService.HandleDriverRejected(ctx, envelope, driverID)

	case "driver.timeout.processed":
		var data struct {
			DriverID string `json:"driver_id"`
		}
		if err := decodeEventData(envelope, &data); err != nil {
			return err
		}

		driverID, err := parseUUID(data.DriverID, "driver_id")
		if err != nil {
			return err
		}

		observability.DriverTimeoutsTotal.WithLabelValues("raw_timeout").Inc()

		return r.driverMetricsService.HandleDriverTimeout(ctx, envelope, driverID)

	case "driver.push_token.updated":
		var data struct {
			DriverID string `json:"driver_id"`
			Token    string `json:"token"`
		}
		if err := decodeEventData(envelope, &data); err != nil {
			return err
		}

		driverID, err := parseUUID(data.DriverID, "driver_id")
		if err != nil {
			return err
		}

		return r.driverCache.AddPushToken(ctx, driverID, data.Token)

	default:
		r.logger.Warn("unhandled event type",
			zap.String("event_type", envelope.Type),
			zap.String("event_id", envelope.ID),
		)
		return nil
	}
}
