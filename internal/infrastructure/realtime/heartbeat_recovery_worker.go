package realtime

import (
	"context"
	"time"

	"github.com/ashadashraf/ride-hail-app/internal/domain/driver"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/observability"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/redis"
	"github.com/ashadashraf/ride-hail-app/internal/ports"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

type HeartbeatRecoveryWorker struct {
	driverCache *redis.DriverCache
	driverCmd   ports.DriverOfflineCommander
	interval    time.Duration
	logger      *zap.Logger
}

func NewHeartbeatRecoveryWorker(
	driverCache *redis.DriverCache,
	driverCmd ports.DriverOfflineCommander,
	interval time.Duration,
	logger *zap.Logger,
) *HeartbeatRecoveryWorker {
	return &HeartbeatRecoveryWorker{
		driverCache: driverCache,
		driverCmd:   driverCmd,
		interval:    interval,
		logger:      logger,
	}
}

var heartbeatRecoveryTracer = otel.Tracer("realtime.heartbeat_recovery")

func (w *HeartbeatRecoveryWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.process(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (w *HeartbeatRecoveryWorker) process(ctx context.Context) {
	ctx, span := heartbeatRecoveryTracer.Start(ctx, "HeartbeatRecoveryWorker.process")
	defer span.End()

	start := time.Now()

	drivers, err := w.driverCache.GetOnlineDriverIDs(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed_to_get_online_drivers")

		w.logger.Error("heartbeat recovery failed to get online drivers",
			zap.Error(err),
		)

		observability.HeartbeatRecoveryScansTotal.WithLabelValues("error").Inc()
		return
	}

	span.SetAttributes(attribute.Int("drivers.online_count", len(drivers)))

	staleCount := 0
	recoveredCount := 0
	failedCount := 0

	for _, driverID := range drivers {
		hasHeartbeat, err := w.driverCache.HasHeartbeat(ctx, driverID)
		if err != nil {
			failedCount++

			w.logger.Warn("heartbeat recovery check failed",
				zap.String("driver_id", driverID.String()),
				zap.Error(err),
			)

			continue
		}

		if hasHeartbeat {
			continue
		}

		staleCount++

		w.logger.Warn("stale driver heartbeat detected",
			zap.String("driver_id", driverID.String()),
			zap.String("reason", string(driver.DriverOfflineReasonHeartbeatExpired)),
		)

		if err := w.driverCmd.GoOffline(
			ctx,
			driverID,
			driver.DriverOfflineReasonHeartbeatExpired,
		); err != nil {
			failedCount++

			span.RecordError(err)

			w.logger.Error("heartbeat recovery failed to mark driver offline",
				zap.String("driver_id", driverID.String()),
				zap.String("reason", string(driver.DriverOfflineReasonHeartbeatExpired)),
				zap.Error(err),
			)

			observability.HeartbeatRecoveriesTotal.WithLabelValues("error").Inc()
			continue
		}

		recoveredCount++

		observability.HeartbeatRecoveriesTotal.WithLabelValues("success").Inc()

		w.logger.Info("heartbeat recovery marked driver offline",
			zap.String("driver_id", driverID.String()),
			zap.String("reason", string(driver.DriverOfflineReasonHeartbeatExpired)),
		)
	}

	duration := time.Since(start)

	span.SetAttributes(
		attribute.Int("drivers.stale_count", staleCount),
		attribute.Int("drivers.recovered_count", recoveredCount),
		attribute.Int("drivers.failed_count", failedCount),
		attribute.Float64("heartbeat_recovery.duration_seconds", duration.Seconds()),
	)

	span.SetStatus(codes.Ok, "heartbeat recovery completed")

	observability.HeartbeatRecoveryScansTotal.WithLabelValues("success").Inc()
	observability.HeartbeatRecoveryDurationSeconds.Observe(duration.Seconds())

	w.logger.Info("heartbeat recovery scan completed",
		zap.Int("online_drivers", len(drivers)),
		zap.Int("stale_drivers", staleCount),
		zap.Int("recovered_drivers", recoveredCount),
		zap.Int("failed_drivers", failedCount),
		zap.Duration("duration", duration),
	)
}
