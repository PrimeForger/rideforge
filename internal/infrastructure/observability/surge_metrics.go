package observability

import "github.com/ashadashraf/ride-hail-app/internal/ports"

type SurgeMetrics struct{}

var _ ports.SurgeObserver = SurgeMetrics{}

func (SurgeMetrics) ObserveSurgeDecision(result string) {
	SurgeDecisionsTotal.WithLabelValues(result).Inc()
}

func (SurgeMetrics) ObserveSurgePolicyRefresh(result string) {
	SurgePolicyRefreshesTotal.WithLabelValues(result).Inc()
}
