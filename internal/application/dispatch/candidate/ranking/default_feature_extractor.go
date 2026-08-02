package ranking

import (
	"context"

	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/candidate"
	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/candidate/pipeline"
	"github.com/ashadashraf/ride-hail-app/internal/shared/geo"
)

type DefaultFeatureExtractor struct{}

func NewDefaultFeatureExtractor() *DefaultFeatureExtractor {
	return &DefaultFeatureExtractor{}
}

func (e *DefaultFeatureExtractor) Extract(
	ctx context.Context,
	pipelineCtx *pipeline.Context,
	c *candidate.Candidate,
) (Features, error) {

	d := c.Driver

	distance := geo.HaversineDistanceKm(
		pipelineCtx.PickupLat,
		pipelineCtx.PickupLng,
		d.Lat,
		d.Lng,
	)

	return Features{
		DistanceKm: distance,

		AcceptanceRate:   d.AcceptanceRate,
		CancellationRate: d.CancellationRate,
		TimeoutRate:      d.TimeoutRate,

		Rating: d.Rating,

		CompletedTrips: d.CompletedRides,

		LastAssignedAt: d.LastAssignedAt,
	}, nil
}
