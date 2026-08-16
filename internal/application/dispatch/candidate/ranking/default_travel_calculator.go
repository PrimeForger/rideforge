package ranking

import (
	"context"
	"errors"

	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/routing"
	"github.com/ashadashraf/ride-hail-app/internal/domain/driver"
	"github.com/ashadashraf/ride-hail-app/internal/shared/geo"
)

type DefaultTravelCalculator struct {
	provider routing.Provider
}

func NewDefaultTravelCalculator(provider routing.Provider) *DefaultTravelCalculator {
	return &DefaultTravelCalculator{
		provider: provider,
	}
}

func (tc *DefaultTravelCalculator) Calculate(
	ctx context.Context,
	pickupLat float64,
	pickupLng float64,
	d *driver.Driver,
) (TravelFeatures, error) {
	if d == nil {
		return TravelFeatures{}, errors.New("driver is nil")
	}

	if tc.provider != nil {
		req := routing.RouteRequest{
			OriginLat:      pickupLat,
			OriginLng:      pickupLng,
			DestinationLat: d.Lat,
			DestinationLng: d.Lng,
		}

		res, err := tc.provider.CalculateRoute(ctx, req)
		if err == nil {
			return TravelFeatures{
				DistanceKm: res.DistanceKm,
				ETASeconds: res.ETASeconds,
			}, nil
		}

		// Context cancellation/timeout must be propagated directly without falling back
		if ctx.Err() != nil {
			return TravelFeatures{}, ctx.Err()
		}

		// On non-context provider error or disabled provider, fall back to Haversine
	}

	distKm := geo.HaversineDistanceKm(
		pickupLat,
		pickupLng,
		d.Lat,
		d.Lng,
	)

	// Note: ETASeconds is set to 0 (explicitly unavailable) when routing is unavailable or disabled.
	return TravelFeatures{
		DistanceKm: distKm,
		ETASeconds: 0,
	}, nil
}
