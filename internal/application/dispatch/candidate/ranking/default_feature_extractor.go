package ranking

import (
	"context"
	"errors"

	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/candidate"
	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/candidate/pipeline"
)

type DefaultFeatureExtractor struct {
	travelCalc TravelCalculator
	enrichers  []Enricher
}

func NewDefaultFeatureExtractor(
	travelCalc TravelCalculator,
	enrichers ...Enricher,
) *DefaultFeatureExtractor {

	return &DefaultFeatureExtractor{
		travelCalc: travelCalc,
		enrichers:  enrichers,
	}
}

func (e *DefaultFeatureExtractor) Extract(
	ctx context.Context,
	pipelineCtx *pipeline.Context,
	c *candidate.Candidate,
) (Features, error) {

	if pipelineCtx == nil {
		return Features{}, errors.New("pipeline context is nil")
	}

	if c == nil {
		return Features{}, errors.New("candidate is nil")
	}

	if c.Driver == nil {
		return Features{}, errors.New("driver is nil")
	}

	d := c.Driver

	travelFeatures, err := e.travelCalc.Calculate(
		ctx,
		pipelineCtx.PickupLat,
		pipelineCtx.PickupLng,
		d,
	)
	if err != nil {
		return Features{}, err
	}

	features := Features{
		Travel: travelFeatures,

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
