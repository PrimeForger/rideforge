package ranking

import (
	"context"

	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/candidate"
	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/candidate/pipeline"
	"github.com/ashadashraf/ride-hail-app/internal/shared/geo"
)

type DefaultFeatureExtractor struct {
	enrichers []Enricher
}

func NewDefaultFeatureExtractor(
	enrichers ...Enricher,
) *DefaultFeatureExtractor {

	return &DefaultFeatureExtractor{
		enrichers: enrichers,
	}
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

	features := Features{
		Travel: TravelFeatures{
			DistanceKm: distance,
			ETASeconds: 0,
		},

		Quality: QualityFeatures{
			AcceptanceRate:   d.AcceptanceRate,
			CancellationRate: d.CancellationRate,
			TimeoutRate:      d.TimeoutRate,
			Rating:           d.Rating,
			CompletedTrips:   d.CompletedRides,
		},

		Fairness: FairnessFeatures{
			LastAssignedAt: d.LastAssignedAt,
		},
	}

	for _, enricher := range e.enrichers {
		if err := enricher.Enrich(
			ctx,
			pipelineCtx,
			c,
			&features,
		); err != nil {
			return Features{}, err
		}
	}

	return features, nil
}
