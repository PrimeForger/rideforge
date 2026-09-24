package observability

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	MatchingAttemptsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "matching_attempts_total",
			Help: "Total matching attempts",
		},
		[]string{"result"},
	)

	DriverOffersTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "driver_offers_total",
			Help: "Total driver offers",
		},
		[]string{"delivery_status"},
	)

	DriverResponsesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "driver_responses_total",
			Help: "Total driver responses",
		},
		[]string{"type"},
	)

	DriverTimeoutsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "driver_timeouts_total",
			Help: "Total driver offer timeouts",
		},
		[]string{"reason"},
	)

	MatchingDurationSeconds = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "matching_duration_seconds",
			Help:    "Matching engine duration",
			Buckets: prometheus.DefBuckets,
		},
	)

	H3CellUpdatesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "h3_cell_updates_total",
			Help: "Total H3 driver cell updates",
		},
		[]string{"result"},
	)

	H3DriverRemovalsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "h3_driver_removals_total",
			Help: "Total H3 driver removals",
		},
		[]string{"result"},
	)

	H3LookupDurationSeconds = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "h3_lookup_duration_seconds",
			Help:    "H3 driver lookup duration",
			Buckets: prometheus.DefBuckets,
		},
	)

	H3LookupResultsTotal = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name: "h3_lookup_result_count",
			Help: "Drivers returned by H3 lookup",
			Buckets: []float64{
				0,
				1,
				5,
				10,
				20,
				50,
				100,
				200,
			},
		},
	)

	H3LookupCellsTotal = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name: "h3_lookup_cell_count",
			Help: "Number of H3 cells searched",
			Buckets: []float64{
				1,
				7,
				19,
				37,
				61,
				91,
			},
		},
	)

	H3IndexedDrivers = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "h3_indexed_drivers",
			Help: "Current number of indexed drivers",
		},
	)

	RedisOperationsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "redis_operations_total",
			Help: "Redis operation count",
		},
		[]string{"operation", "result"},
	)

	KafkaEventsProcessedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kafka_events_processed_total",
			Help: "Kafka events processed",
		},
		[]string{"event_type", "result"},
	)

	HeartbeatRecoveryScansTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "heartbeat_recovery_scans_total",
			Help: "Total heartbeat recovery scans",
		},
		[]string{"result"},
	)

	HeartbeatRecoveriesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "heartbeat_recoveries_total",
			Help: "Total stale driver heartbeat recoveries",
		},
		[]string{"result"},
	)

	HeartbeatRecoveryDurationSeconds = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "heartbeat_recovery_duration_seconds",
			Help:    "Heartbeat recovery scan duration",
			Buckets: prometheus.DefBuckets,
		},
	)

	// Stage 10 / Phase B — Dispatch Latency Metrics

	DispatchCandidateLookupDurationSeconds = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "dispatch_candidate_lookup_duration_seconds",
			Help:    "Candidate lookup duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0},
		},
	)

	DispatchRankingDurationSeconds = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "dispatch_ranking_duration_seconds",
			Help:    "Candidate pipeline ranking duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0},
		},
	)

	DispatchReservationDurationSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "dispatch_reservation_duration_seconds",
			Help:    "Driver reservation lock duration in seconds",
			Buckets: []float64{0.0005, 0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0},
		},
		[]string{"result"},
	)

	DispatchOfferDurationSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "dispatch_offer_duration_seconds",
			Help:    "Driver offer delivery duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0},
		},
		[]string{"status"},
	)

	DispatchAcceptanceDurationSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "dispatch_acceptance_duration_seconds",
			Help:    "Driver acceptance and response handling duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0},
		},
		[]string{"result"},
	)

	// Stage 10 / Phase C — H3 & Candidate Retrieval Metrics

	H3SearchRingsVisited = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "h3_search_rings_visited",
			Help:    "Number of H3 rings visited during candidate discovery",
			Buckets: []float64{0, 1, 2, 3, 4, 5, 8, 10},
		},
	)

	CandidatePipelineCount = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "candidate_pipeline_count",
			Help:    "Candidate count at candidate pipeline processing stages",
			Buckets: []float64{0, 1, 5, 10, 20, 50, 100, 200},
		},
		[]string{"stage"},
	)

	CandidateRetrievalOutcomeTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "candidate_retrieval_outcome_total",
			Help: "Candidate discovery and retrieval outcomes",
		},
		[]string{"result"},
	)

	// Stage 10 / Phase D — Operational Spatial / Heatmap Metrics

	DispatchDensityClassificationsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "dispatch_density_classifications_total",
			Help: "Total H3 ring driver density classifications evaluated during search expansion",
		},
		[]string{"density"},
	)

	// Stage 7 / Phase A — Cell Demand Metrics

	CellDemandEventsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cell_demand_events_total",
			Help: "Total cell demand events processed",
		},
		[]string{"event_type", "result"},
	)

	CellActiveDemandTotal = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "cell_active_demand_total",
			Help: "Current total active ride requests across all cells",
		},
	)

	DemandReconciliationDurationSeconds = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "demand_reconciliation_duration_seconds",
			Help:    "Demand reconciliation execution duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
	)

	// Stage 7 / Phase B — Hot Cell & Occupancy Metrics

	HotCellClassificationsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "hot_cell_classifications_total",
			Help: "Total H3 cell occupancy classifications by status",
		},
		[]string{"status"},
	)

	CellImbalanceRatio = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "cell_imbalance_ratio",
			Help:    "Distribution of calculated demand-to-supply imbalance ratios",
			Buckets: []float64{0, 0.5, 1.0, 1.2, 1.5, 2.0, 3.0, 5.0, 10.0},
		},
	)

	HotCellsActiveTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "hot_cells_active_total",
			Help: "Current count of active hot and warm cells",
		},
		[]string{"status"},
	)
)

var (
	collectors = []prometheus.Collector{
		MatchingAttemptsTotal,
		DriverOffersTotal,
		DriverResponsesTotal,
		DriverTimeoutsTotal,
		MatchingDurationSeconds,
		H3CellUpdatesTotal,
		H3DriverRemovalsTotal,
		H3LookupDurationSeconds,
		H3LookupResultsTotal,
		H3LookupCellsTotal,
		H3IndexedDrivers,
		RedisOperationsTotal,
		KafkaEventsProcessedTotal,
		HeartbeatRecoveryScansTotal,
		HeartbeatRecoveriesTotal,
		HeartbeatRecoveryDurationSeconds,
		DispatchCandidateLookupDurationSeconds,
		DispatchRankingDurationSeconds,
		DispatchReservationDurationSeconds,
		DispatchOfferDurationSeconds,
		DispatchAcceptanceDurationSeconds,
		H3SearchRingsVisited,
		CandidatePipelineCount,
		CandidateRetrievalOutcomeTotal,
		DispatchDensityClassificationsTotal,
		CellDemandEventsTotal,
		CellActiveDemandTotal,
		DemandReconciliationDurationSeconds,
		HotCellClassificationsTotal,
		CellImbalanceRatio,
		HotCellsActiveTotal,
	}

	registerOnce sync.Once
)

// Register registers all observability metrics with the default Prometheus registerer.
func Register() {
	registerOnce.Do(func() {
		RegisterWithRegistry(prometheus.DefaultRegisterer)
	})
}

// RegisterWithRegistry registers all observability metrics with the provided Prometheus registerer.
func RegisterWithRegistry(reg prometheus.Registerer) {
	for _, c := range collectors {
		_ = reg.Register(c)
	}
}
