package ranking

import (
	"context"
	"math"
	"time"

	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/candidate"
	"github.com/ashadashraf/ride-hail-app/internal/config"
)

type DefaultScorer struct {
	cfg *config.RankingConfig
}

func NewDefaultScorer(cfg *config.RankingConfig) *DefaultScorer {
	return &DefaultScorer{
		cfg: cfg,
	}
}

func (s *DefaultScorer) Score(
	ctx context.Context,
	f Features,
) (candidate.Score, error) {

	// Normalize inputs (avoid dominance)
	distanceScore := 1 / (1 + f.DistanceKm)

	acceptanceScore := f.AcceptanceRate
	cancellationPenalty := 1 - f.CancellationRate
	timeoutPenalty := 1 - f.TimeoutRate

	ratingScore := f.Rating / 5.0

	experienceScore := math.Log(float64(f.CompletedTrips) + 1)

	// fairness boost (older idle drivers get slight boost)
	fairnessScore := time.Since(f.LastAssignedAt).Seconds() / 3600

	// Final weighted score
	finalScore := (s.cfg.DistanceWeight * distanceScore) +
		(s.cfg.AcceptanceWeight * acceptanceScore) +
		(s.cfg.CancellationWeight * cancellationPenalty) +
		(s.cfg.TimeoutWeight * timeoutPenalty) +
		(s.cfg.RatingWeight * ratingScore) +
		(s.cfg.ExperienceWeight * experienceScore) +
		(s.cfg.FairnessWeight * fairnessScore)

	return candidate.Score{
		Value: finalScore,
	}, nil
}
