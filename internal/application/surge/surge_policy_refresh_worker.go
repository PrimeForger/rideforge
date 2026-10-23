package surge

import (
	"context"
	"time"

	"github.com/ashadashraf/ride-hail-app/internal/ports"
	"go.uber.org/zap"
)

type SurgePolicyRefresher interface {
	Refresh(ctx context.Context) error
}

type SurgePolicyRefreshWorker struct {
	refresher SurgePolicyRefresher
	interval  time.Duration
	observer  ports.SurgeObserver
	log       *zap.Logger
}

func NewSurgePolicyRefreshWorker(refresher SurgePolicyRefresher, interval time.Duration, observer ports.SurgeObserver, log *zap.Logger) *SurgePolicyRefreshWorker {
	return &SurgePolicyRefreshWorker{refresher: refresher, interval: interval, observer: observer, log: log}
}

func (w *SurgePolicyRefreshWorker) Start(ctx context.Context) {
	if w.interval <= 0 {
		if w.log != nil {
			w.log.Error("surge policy refresh worker not started: interval must be positive")
		}
		return
	}
	w.refresh(ctx)
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			w.refresh(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (w *SurgePolicyRefreshWorker) refresh(ctx context.Context) {
	if err := w.refresher.Refresh(ctx); err != nil {
		if w.observer != nil {
			w.observer.ObserveSurgePolicyRefresh("error")
		}
		if w.log != nil {
			w.log.Error("surge policy refresh failed; quotes will use base fare when configuration is unavailable", zap.Error(err))
		}
		return
	}
	if w.observer != nil {
		w.observer.ObserveSurgePolicyRefresh("success")
	}
}
