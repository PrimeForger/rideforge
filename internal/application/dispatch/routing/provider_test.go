package routing_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/routing"
)

func TestDisabledProvider_CalculateRoute(t *testing.T) {
	provider := routing.NewDisabledProvider()
	ctx := context.Background()

	req := routing.RouteRequest{
		OriginLat:      12.9716,
		OriginLng:      77.5946,
		DestinationLat: 12.9750,
		DestinationLng: 77.5990,
	}

	res, err := provider.CalculateRoute(ctx, req)
	if err == nil {
		t.Fatalf("expected error from DisabledProvider, got nil")
	}

	if !errors.Is(err, routing.ErrProviderDisabled) {
		t.Errorf("expected ErrProviderDisabled, got %v", err)
	}

	if res.DistanceKm != 0 || res.ETASeconds != 0 {
		t.Errorf("expected zero RouteResult, got %+v", res)
	}
}
