package application

import (
	"context"
	"database/sql"
	"errors"

	appevents "github.com/ashadashraf/ride-hail-app/internal/application/events"
	"github.com/ashadashraf/ride-hail-app/internal/application/matching"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/observability"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/postgres"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/redis"
	"github.com/google/uuid"
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
		driverLocker:          driverLocker,
		logger:                logger,
	}
}

func (r *EventRouter) Handle(ctx context.Context, envelope appevents.Envelope) error {
	if envelope.ID == "" {
		return errors.New("event id is required")
	}

	if envelope.Type == "" {
		return errors.New("event type is required")
	}

	switch envelope.Type {
	case "ride.requested", "matching.started", "matching.retry",
		"driver.accepted", "driver.rejected", "driver.timeout":

		return r.handleTransactional(ctx, envelope)

	default:
		return r.handleSideEffect(ctx, envelope)
	}
}

func (r *EventRouter) handleTransactional(
	ctx context.Context,
	envelope appevents.Envelope,
) error {
	return r.txManager.WithinTx(ctx, func(tx *sql.Tx) error {

		eventID, err := uuid.Parse(envelope.ID)
		if err != nil {
			return err
		}

		processed, err := r.processedEventRepo.InsertIfNotExistsTx(
			ctx,
			tx,
			eventID,
			"matching-service",
		)
		if err != nil {
			return err
		}

		if !processed {
			r.logger.Info(
				"duplicate event skipped",
				zap.String("event_id", envelope.ID),
				zap.String("event_type", envelope.Type),
			)
			return nil
		}

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
	})
}

func (r *EventRouter) handleSideEffect(
	ctx context.Context,
	envelope appevents.Envelope,
) error {

	eventID, err := uuid.Parse(envelope.ID)
	if err != nil {
		return err
	}

	processed, err := r.processedEventRepo.InsertIfNotExists(
		ctx,
		eventID,
		"matching-service",
	)
	if err != nil {
		return err
	}

	if !processed {
		r.logger.Info(
			"duplicate event skipped",
			zap.String("event_id", envelope.ID),
			zap.String("event_type", envelope.Type),
		)
		return nil
	}

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

		if err := r.driverCache.MarkOffline(ctx, driverID); err != nil {
			return err
		}

		return r.driverLocker.ForceRelease(ctx, driverID)

	case "driver.offered":
		var data driverRideEventData
		if err := decodeEventData(envelope, &data); err != nil {
			return err
		}

		rideID, driverID, err := parseRideDriverIDs(data)
		if err != nil {
			return err
		}

		if err := r.driverOfferHandler.HandleDriverOffered(ctx, rideID, driverID); err != nil {
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
