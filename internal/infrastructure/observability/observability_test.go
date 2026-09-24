package observability_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/observability"
	"github.com/prometheus/client_golang/prometheus"
)

func TestMetricsRegistration(t *testing.T) {
	reg := prometheus.NewRegistry()
	observability.RegisterWithRegistry(reg)

	// Exercise legacy metrics
	observability.MatchingAttemptsTotal.WithLabelValues("success").Inc()
	observability.DriverOffersTotal.WithLabelValues("delivered_ws").Inc()
	observability.DriverResponsesTotal.WithLabelValues("accepted").Inc()
	observability.DriverTimeoutsTotal.WithLabelValues("raw_timeout").Inc()
	observability.MatchingDurationSeconds.Observe(0.045)

	observability.H3CellUpdatesTotal.WithLabelValues("success").Inc()
	observability.H3DriverRemovalsTotal.WithLabelValues("success").Inc()
	observability.H3LookupDurationSeconds.Observe(0.002)
	observability.H3LookupResultsTotal.Observe(5)
	observability.H3LookupCellsTotal.Observe(7)
	observability.H3IndexedDrivers.Set(42)

	observability.RedisOperationsTotal.WithLabelValues("geo_radius", "ok").Inc()
	observability.KafkaEventsProcessedTotal.WithLabelValues("ride_created", "success").Inc()
	observability.HeartbeatRecoveryScansTotal.WithLabelValues("success").Inc()
	observability.HeartbeatRecoveriesTotal.WithLabelValues("success").Inc()
	observability.HeartbeatRecoveryDurationSeconds.Observe(0.012)

	// Exercise Stage 10 Phase B Latency Metrics
	observability.DispatchCandidateLookupDurationSeconds.Observe(0.005)
	observability.DispatchRankingDurationSeconds.Observe(0.012)
	observability.DispatchReservationDurationSeconds.WithLabelValues("success").Observe(0.003)
	observability.DispatchOfferDurationSeconds.WithLabelValues("delivered_ws").Observe(0.015)
	observability.DispatchAcceptanceDurationSeconds.WithLabelValues("accepted").Observe(0.250)

	// Exercise Stage 10 Phase C H3 & Candidate Retrieval Metrics
	observability.H3SearchRingsVisited.Observe(2)
	observability.CandidatePipelineCount.WithLabelValues("loaded").Observe(10)
	observability.CandidatePipelineCount.WithLabelValues("filtered").Observe(2)
	observability.CandidatePipelineCount.WithLabelValues("ranked").Observe(8)
	observability.CandidateRetrievalOutcomeTotal.WithLabelValues("found").Inc()

	// Exercise Stage 10 Phase D Operational Spatial / Heatmap Metrics
	observability.DispatchDensityClassificationsTotal.WithLabelValues("normal").Inc()

	metricFamilies, err := reg.Gather()
	if err != nil {
		t.Fatalf("failed to gather registered metrics: %v", err)
	}

	if len(metricFamilies) == 0 {
		t.Fatal("expected gathered metric families to be non-empty")
	}

	expectedMetrics := map[string]bool{
		"matching_attempts_total":                    false,
		"driver_offers_total":                        false,
		"matching_duration_seconds":                  false,
		"h3_indexed_drivers":                         false,
		"dispatch_candidate_lookup_duration_seconds": false,
		"dispatch_ranking_duration_seconds":          false,
		"dispatch_reservation_duration_seconds":      false,
		"dispatch_offer_duration_seconds":            false,
		"dispatch_acceptance_duration_seconds":       false,
		"h3_search_rings_visited":                    false,
		"candidate_pipeline_count":                   false,
		"candidate_retrieval_outcome_total":          false,
		"dispatch_density_classifications_total":     false,
	}

	for _, mf := range metricFamilies {
		name := mf.GetName()
		if _, ok := expectedMetrics[name]; ok {
			expectedMetrics[name] = true
		}
	}

	for name, found := range expectedMetrics {
		if !found {
			t.Errorf("expected metric family %q to be gathered", name)
		}
	}
}

func TestMetricCardinalityPolicy(t *testing.T) {
	reg := prometheus.NewRegistry()
	observability.RegisterWithRegistry(reg)

	metricFamilies, err := reg.Gather()
	if err != nil {
		t.Fatalf("failed to gather registered metrics: %v", err)
	}

	disallowedLabels := []string{
		"ride_id", "driver_id", "user_id", "request_id",
		"cell_id", "h3_cell", "latitude", "longitude", "lat", "lng",
		"error_message",
	}

	for _, mf := range metricFamilies {
		for _, m := range mf.GetMetric() {
			for _, label := range m.GetLabel() {
				labelName := strings.ToLower(label.GetName())
				for _, disallowed := range disallowedLabels {
					if labelName == disallowed {
						t.Errorf("cardinality violation in metric %s: forbidden label name %q found", mf.GetName(), labelName)
					}
				}
			}
		}
	}
}

func TestInitTracing_EmptyEndpoint(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	shutdown, err := observability.InitTracing(ctx, "test-service", "test", "")
	if err != nil {
		t.Fatalf("expected no error when OTLP endpoint is empty, got: %v", err)
	}

	if shutdown == nil {
		t.Fatal("expected non-nil shutdown function")
	}

	if err := shutdown(ctx); err != nil {
		t.Errorf("expected clean shutdown execution, got: %v", err)
	}
}

func TestNoopShutdown(t *testing.T) {
	err := observability.NoopShutdown(context.Background())
	if err != nil {
		t.Errorf("expected NoopShutdown to return nil, got: %v", err)
	}
}
