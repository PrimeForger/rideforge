package application

import (
	"context"
	"database/sql"

	appevents "github.com/ashadashraf/ride-hail-app/internal/application/events"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/postgres"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/redis"
	"github.com/ashadashraf/ride-hail-app/internal/ports"
	"github.com/google/uuid"
)

type DriverMetricsService struct {
	txManager   *postgres.TxManager
	metricsRepo ports.DriverMetricsRepository
	driverCache *redis.DriverCache
}

func NewDriverMetricsService(
	txManager *postgres.TxManager,
	metricsRepo ports.DriverMetricsRepository,
	driverCache *redis.DriverCache,
) *DriverMetricsService {
	return &DriverMetricsService{
		txManager:   txManager,
		metricsRepo: metricsRepo,
		driverCache: driverCache,
	}
}

func (s *DriverMetricsService) HandleEvent(
	ctx context.Context,
	envelope appevents.Envelope,
	driverID uuid.UUID,
	apply func(ctx context.Context, tx *sql.Tx) error,
) error {
	eventID, err := uuid.Parse(envelope.ID)
	if err != nil {
		return err
	}

	return s.txManager.WithinTx(ctx, func(tx *sql.Tx) error {
		inserted, err := s.metricsRepo.InsertMetricEventTx(
			ctx,
			tx,
			eventID,
			driverID,
			envelope.Type,
		)
		if err != nil {
			return err
		}

		if !inserted {
			return nil
		}

		if err := apply(ctx, tx); err != nil {
			return err
		}

		if err := s.metricsRepo.RecalculateRatesTx(ctx, tx, driverID); err != nil {
			return err
		}

		snapshot, err := s.metricsRepo.GetMetricsTx(ctx, tx, driverID)
		if err != nil {
			return err
		}

		return s.driverCache.UpdateDriverMetrics(
			ctx,
			driverID,
			snapshot.AcceptanceRate,
			snapshot.CancellationRate,
			snapshot.TimeoutRate,
			snapshot.CompletedRides,
		)
	})
}

func (s *DriverMetricsService) HandleDriverOffered(
	ctx context.Context,
	envelope appevents.Envelope,
	driverID uuid.UUID,
) error {
	return s.HandleEvent(ctx, envelope, driverID, func(ctx context.Context, tx *sql.Tx) error {
		return s.metricsRepo.IncrementOfferedTx(ctx, tx, driverID)
	})
}

func (s *DriverMetricsService) HandleDriverOfferAcked(
	ctx context.Context,
	envelope appevents.Envelope,
	driverID uuid.UUID,
) error {
	return s.HandleEvent(ctx, envelope, driverID, func(ctx context.Context, tx *sql.Tx) error {
		return s.metricsRepo.IncrementAckedTx(ctx, tx, driverID)
	})
}

func (s *DriverMetricsService) HandleDriverAccepted(
	ctx context.Context,
	envelope appevents.Envelope,
	driverID uuid.UUID,
) error {
	return s.HandleEvent(ctx, envelope, driverID, func(ctx context.Context, tx *sql.Tx) error {
		return s.metricsRepo.IncrementAcceptedTx(ctx, tx, driverID)
	})
}

func (s *DriverMetricsService) HandleDriverRejected(
	ctx context.Context,
	envelope appevents.Envelope,
	driverID uuid.UUID,
) error {
	return s.HandleEvent(ctx, envelope, driverID, func(ctx context.Context, tx *sql.Tx) error {
		return s.metricsRepo.IncrementRejectedTx(ctx, tx, driverID)
	})
}

func (s *DriverMetricsService) HandleDriverTimeout(
	ctx context.Context,
	envelope appevents.Envelope,
	driverID uuid.UUID,
) error {
	return s.HandleEvent(ctx, envelope, driverID, func(ctx context.Context, tx *sql.Tx) error {
		return s.metricsRepo.IncrementTimeoutTx(ctx, tx, driverID)
	})
}

func (s *DriverMetricsService) HandleRideCompleted(
	ctx context.Context,
	envelope appevents.Envelope,
	driverID uuid.UUID,
) error {
	return s.HandleEvent(ctx, envelope, driverID, func(ctx context.Context, tx *sql.Tx) error {
		return s.metricsRepo.IncrementCompletedTx(ctx, tx, driverID)
	})
}

func (s *DriverMetricsService) HandleDriverCancelled(
	ctx context.Context,
	envelope appevents.Envelope,
	driverID uuid.UUID,
) error {
	return s.HandleEvent(ctx, envelope, driverID, func(ctx context.Context, tx *sql.Tx) error {
		return s.metricsRepo.IncrementCancelledTx(ctx, tx, driverID)
	})
}
