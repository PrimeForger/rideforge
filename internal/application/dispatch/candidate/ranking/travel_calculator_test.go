package ranking_test

import (
	"context"
	"errors"
	"math"
	"testing"

	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/candidate/ranking"
	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/routing"
	"github.com/ashadashraf/ride-hail-app/internal/domain/driver"
	"github.com/ashadashraf/ride-hail-app/internal/shared/geo"
	"github.com/google/uuid"
)

type testRoutingProvider struct {
	result routing.RouteResult
	err    error
}

func (p *testRoutingProvider) CalculateRoute(ctx context.Context, req routing.RouteRequest) (routing.RouteResult, error) {
	if ctx.Err() != nil {
		return routing.RouteResult{}, ctx.Err()
	}
	return p.result, p.err
}

func TestDefaultTravelCalculator_Calculate_ProviderSuccess(t *testing.T) {
	provider := &testRoutingProvider{
		result: routing.RouteResult{
			DistanceKm: 4.2,
			ETASeconds: 320,
		},
	}

	calc := ranking.NewDefaultTravelCalculator(provider)
	ctx := context.Background()

	d := &driver.Driver{
		ID:  uuid.New(),
		Lat: 12.9750,
		Lng: 77.5990,
	}

	features, err := calc.Calculate(ctx, 12.9716, 77.5946, d)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if features.DistanceKm != 4.2 {
		t.Errorf("expected DistanceKm 4.2, got %f", features.DistanceKm)
	}
	if features.ETASeconds != 320 {
		t.Errorf("expected ETASeconds 320, got %f", features.ETASeconds)
	}
}

func TestDefaultTravelCalculator_Calculate_ProviderDisabled_HaversineFallback(t *testing.T) {
	provider := routing.NewDisabledProvider()
	calc := ranking.NewDefaultTravelCalculator(provider)
	ctx := context.Background()

	pickupLat, pickupLng := 12.9716, 77.5946
	d := &driver.Driver{
		ID:  uuid.New(),
		Lat: 12.9750,
		Lng: 77.5990,
	}

	features, err := calc.Calculate(ctx, pickupLat, pickupLng, d)
	if err != nil {
		t.Fatalf("expected no error on fallback, got %v", err)
	}

	expectedDist := geo.HaversineDistanceKm(pickupLat, pickupLng, d.Lat, d.Lng)
	if math.Abs(features.DistanceKm-expectedDist) > 0.0001 {
		t.Errorf("expected Haversine distance %f, got %f", expectedDist, features.DistanceKm)
	}
	if features.ETASeconds != 0 {
		t.Errorf("expected ETASeconds 0 on disabled provider fallback, got %f", features.ETASeconds)
	}
}

func TestDefaultTravelCalculator_Calculate_ProviderFailure_HaversineFallback(t *testing.T) {
	provider := &testRoutingProvider{
		err: errors.New("upstream routing service unavailable"),
	}

	calc := ranking.NewDefaultTravelCalculator(provider)
	ctx := context.Background()

	pickupLat, pickupLng := 12.9716, 77.5946
	d := &driver.Driver{
		ID:  uuid.New(),
		Lat: 12.9750,
		Lng: 77.5990,
	}

	features, err := calc.Calculate(ctx, pickupLat, pickupLng, d)
	if err != nil {
		t.Fatalf("expected no error on provider failure fallback, got %v", err)
	}

	expectedDist := geo.HaversineDistanceKm(pickupLat, pickupLng, d.Lat, d.Lng)
	if math.Abs(features.DistanceKm-expectedDist) > 0.0001 {
		t.Errorf("expected Haversine distance %f, got %f", expectedDist, features.DistanceKm)
	}
	if features.ETASeconds != 0 {
		t.Errorf("expected ETASeconds 0 on failure fallback, got %f", features.ETASeconds)
	}
}

func TestDefaultTravelCalculator_Calculate_ContextCancelled_NoFallback(t *testing.T) {
	provider := &testRoutingProvider{
		err: errors.New("should not reach here without context error"),
	}

	calc := ranking.NewDefaultTravelCalculator(provider)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	d := &driver.Driver{
		ID:  uuid.New(),
		Lat: 12.9750,
		Lng: 77.5990,
	}

	_, err := calc.Calculate(ctx, 12.9716, 77.5946, d)
	if err == nil {
		t.Fatalf("expected context cancellation error, got nil")
	}

	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled error, got %v", err)
	}
}

func TestDefaultTravelCalculator_Calculate_NilDriver(t *testing.T) {
	calc := ranking.NewDefaultTravelCalculator(routing.NewDisabledProvider())
	ctx := context.Background()

	_, err := calc.Calculate(ctx, 12.9716, 77.5946, nil)
	if err == nil {
		t.Fatalf("expected error for nil driver, got nil")
	}

	if err.Error() != "driver is nil" {
		t.Errorf("expected 'driver is nil' error, got '%v'", err)
	}
}
