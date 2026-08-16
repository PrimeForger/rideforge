package routing

import (
	"context"
	"errors"
)

var ErrProviderDisabled = errors.New("routing provider is disabled")

type RouteRequest struct {
	OriginLat      float64
	OriginLng      float64
	DestinationLat float64
	DestinationLng float64
}

type RouteResult struct {
	DistanceKm float64
	ETASeconds float64
}

type Provider interface {
	CalculateRoute(ctx context.Context, req RouteRequest) (RouteResult, error)
}
