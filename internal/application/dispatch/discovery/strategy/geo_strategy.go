package strategy

import (
	"context"

	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/discovery/search"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/redis"
	"github.com/google/uuid"
)

type GeoStrategy struct {
	geo *redis.GeoService
}

func NewGeoStrategy(
	geo *redis.GeoService,
) *GeoStrategy {
	return &GeoStrategy{
		geo: geo,
	}
}

func (s *GeoStrategy) FindCandidates(
	ctx context.Context,
	req search.Request,
) (search.Result, error) {

	nearby, err := s.geo.FindNearbyDriversWithDistance(
		ctx,
		req.PickupLat,
		req.PickupLng,
		req.RadiusKm,
		req.CandidateLimit,
	)

	if err != nil {
		return search.Result{}, err
	}

	ids := make([]uuid.UUID, 0, len(nearby))

	for _, d := range nearby {
		ids = append(ids, d.ID)
	}

	return search.Result{
		DriverIDs: ids,

		Backend: "geo",

		RadiusKm: req.RadiusKm,
	}, nil
}
