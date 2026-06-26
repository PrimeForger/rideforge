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
		)
	})
}
