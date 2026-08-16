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

	var travelScore float64
	if travel.ETASeconds > 0 {
		travelScore = s.calculateETAScore(travel.ETASeconds)
	} else {
		travelScore = s.calculateDistanceScore(travel.DistanceKm)
	}

	qualityScore := s.qualityCalculator.Calculate(
		quality,
	)

	// fairness boost (older idle drivers get slight boost)
	var fairnessScore float64
	if !fairness.LastAssignedAt.IsZero() {
		idleSec := time.Since(fairness.LastAssignedAt).Seconds()
		if idleSec > 0 {
			fairnessScore = idleSec / 3600.0
		}
	}

	// Final weighted score
	finalScore :=
		(s.cfg.DistanceWeight * travelScore) +
			(s.cfg.FairnessWeight * fairnessScore) +
			qualityScore

	return candidate.Score{
		Value: finalScore,
	}, nil
}

func (s *DefaultScorer) calculateETAScore(etaSeconds float64) float64 {
	maxETA := s.cfg.MaxETASeconds
	if maxETA <= 0 || etaSeconds >= maxETA || etaSeconds <= 0 {
		return 0.0
	}

	score := 1.0 - (etaSeconds / maxETA)
	if score < 0.0 {
		return 0.0
	}
	if score > 1.0 {
		return 1.0
	}
	return score
}

func (s *DefaultScorer) calculateDistanceScore(distanceKm float64) float64 {
	if distanceKm < 0 {
		return 0.0
	}

	score := 1.0 / (1.0 + distanceKm)
	if score < 0.0 {
		return 0.0
	}
	if score > 1.0 {
		return 1.0
	}
	return score
}
