package application

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/observability"
	"go.uber.org/zap"
)

type DemandRecoveryRunner interface {
	RebuildDemand(ctx context.Context) (DemandRecoveryResult, error)
}

type DemandReconciliationWorker struct {
	recoveryRunner DemandRecoveryRunner
	enabled        bool
	interval       time.Duration
	log            *zap.Logger
	running        int32
}

func NewDemandReconciliationWorker(
	recoveryRunner DemandRecoveryRunner,
	enabled bool,
	interval time.Duration,
	log *zap.Logger,
) *DemandReconciliationWorker {
	return &DemandReconciliationWorker{
		recoveryRunner: recoveryRunner,
		enabled:        enabled,
		interval:       interval,
		log:            log,
	}
}

func (w *DemandReconciliationWorker) Start(ctx context.Context) {
	if !w.enabled {
		if w.log != nil {
			w.log.Info("Demand reconciliation worker is disabled")
		}
		return
	}

	if w.log != nil {
		w.log.Info("Demand reconciliation worker starting",
			zap.Duration("interval", w.interval),
		)
	}

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.runReconciliation(ctx)
		case <-ctx.Done():
			if w.log != nil {
				w.log.Info("Demand reconciliation worker stopping")
			}
			return
		}
	}
}

func (w *DemandReconciliationWorker) runReconciliation(ctx context.Context) {
	if !atomic.CompareAndSwapInt32(&w.running, 0, 1) {
		if w.log != nil {
			w.log.Warn("skipping Demand reconciliation: previous run is still in progress")
		}
		return
	}
	defer atomic.StoreInt32(&w.running, 0)

	start := time.Now()
	result, err := w.recoveryRunner.RebuildDemand(ctx)
	duration := time.Since(start)

	if err != nil {
		observability.DemandReconciliationDurationSeconds.Observe(duration.Seconds())
		if w.log != nil {
			w.log.Error("Demand reconciliation run failed", zap.Error(err))
		}
		return
	}

	observability.DemandReconciliationDurationSeconds.Observe(duration.Seconds())
	observability.CellActiveDemandTotal.Set(float64(result.TotalActiveRides))

	if w.log != nil {
		w.log.Info("Demand reconciliation run completed",
			zap.Int("total_active_rides", result.TotalActiveRides),
			zap.Int("total_cells", result.TotalCells),
			zap.Duration("duration", result.Duration),
		)
	}
}
