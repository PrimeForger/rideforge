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
)

var registerOnce sync.Once

func Register() {
	registerOnce.Do(func() {
		prometheus.MustRegister(
			MatchingAttemptsTotal,
			DriverOffersTotal,
			DriverResponsesTotal,
			DriverTimeoutsTotal,
			MatchingDurationSeconds,
			RedisOperationsTotal,
			KafkaEventsProcessedTotal,
			HeartbeatRecoveryScansTotal,
			HeartbeatRecoveriesTotal,
			HeartbeatRecoveryDurationSeconds,
		)
	})
}
