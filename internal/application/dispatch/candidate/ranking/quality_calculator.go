package ranking

import (
	"math"

	"github.com/ashadashraf/ride-hail-app/internal/config"
)

type QualityCalculator struct {
	cfg *config.RankingConfig
}

func NewQualityCalculator(
	cfg *config.RankingConfig,
) *QualityCalculator {

	return &QualityCalculator{
		cfg: cfg,
	}
}

func (c *QualityCalculator) Calculate(
	quality QualityFeatures,
) float64 {

	acceptanceScore := quality.AcceptanceRate

	cancellationPenalty := 1 - quality.CancellationRate

	timeoutPenalty := 1 - quality.TimeoutRate

	ratingScore := quality.Rating / 5.0

	experienceScore := math.Log(
		float64(quality.CompletedTrips) + 1,
	)

	return (c.cfg.AcceptanceWeight * acceptanceScore) +
		(c.cfg.CancellationWeight * cancellationPenalty) +
		(c.cfg.TimeoutWeight * timeoutPenalty) +
		(c.cfg.RatingWeight * ratingScore) +
		(c.cfg.ExperienceWeight * experienceScore)
}
