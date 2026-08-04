package ranking

import (
	"context"
	"time"

	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/candidate"
	"github.com/ashadashraf/ride-hail-app/internal/config"
)

type DefaultScorer struct {
	cfg *config.RankingConfig

	qualityCalculator *QualityCalculator
}

func NewDefaultScorer(
	cfg *config.RankingConfig,
) *DefaultScorer {

	return &DefaultScorer{
		cfg: cfg,

		qualityCalculator: NewQualityCalculator(cfg),
	}
}

func (s *DefaultScorer) Score(
	ctx context.Context,
	f Features,
) (candidate.Score, error) {

	travel := f.Travel
	quality := f.Quality
	fairness := f.Fairness

	// Normalize inputs (avoid dominance)
	distanceScore := 1 / (1 + travel.DistanceKm)

	qualityScore := s.qualityCalculator.Calculate(
		quality,
	)

	// fairness boost (older idle drivers get slight boost)
	fairnessScore := time.Since(fairness.LastAssignedAt).Seconds() / 3600

	// Final weighted score
	finalScore :=
		(s.cfg.DistanceWeight * distanceScore) +
			(s.cfg.FairnessWeight * fairnessScore) +
			qualityScore

	return candidate.Score{
		Value: finalScore,
	}, nil
}
