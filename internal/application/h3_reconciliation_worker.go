package application

import (
	"context"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

type RecoveryRunner interface {
	RebuildIndex(ctx context.Context) (RecoveryResult, error)
}

type H3ReconciliationWorker struct {
	recoveryRunner RecoveryRunner
	enabled        bool
	interval       time.Duration
	log            *zap.Logger
	running        int32
}

func NewH3ReconciliationWorker(
	recoveryRunner RecoveryRunner,
	enabled bool,
	interval time.Duration,
	log *zap.Logger,
) *H3ReconciliationWorker {
	return &H3ReconciliationWorker{
		recoveryRunner: recoveryRunner,
		enabled:        enabled,
		interval:       interval,
		log:            log,
	}
}

func (w *H3ReconciliationWorker) Start(ctx context.Context) {
	if !w.enabled {
		if w.log != nil {
			w.log.Info("H3 reconciliation worker is disabled")
		}
		return
	}

	if w.log != nil {
		w.log.Info("H3 reconciliation worker starting",
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
				w.log.Info("H3 reconciliation worker stopping")
			}
			return
		}
	}
}

func (w *H3ReconciliationWorker) runReconciliation(ctx context.Context) {
	if !atomic.CompareAndSwapInt32(&w.running, 0, 1) {
		if w.log != nil {
			w.log.Warn("skipping H3 reconciliation: previous run is still in progress")
		}
		return
	}
	defer atomic.StoreInt32(&w.running, 0)

	result, err := w.recoveryRunner.RebuildIndex(ctx)
	if err != nil {
		if w.log != nil {
			w.log.Error("H3 reconciliation run failed", zap.Error(err))
		}
		return
	}

	if w.log != nil {
		w.log.Info("H3 reconciliation run completed",
			zap.Int("total_drivers", result.TotalDrivers),
			zap.Int("restored_drivers", result.RestoredDrivers),
			zap.Int("skipped_drivers", result.SkippedDrivers),
			zap.Int("cell_updates", result.CellUpdates),
			zap.Duration("duration", result.Duration),
		)
	}
}
