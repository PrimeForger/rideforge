package matching

import (
	"math"
	"time"

	"github.com/ashadashraf/ride-hail-app/internal/config"
	"github.com/ashadashraf/ride-hail-app/internal/domain/driver"
)

type RankingEngine struct {
	cfg *config.RankingConfig
}

func NewRankingEngine(cfg *config.RankingConfig) *RankingEngine {
	return &RankingEngine{
		cfg: cfg,
	}
}

func (r *RankingEngine) Score(
	driver *driver.Driver,
	distanceKm float64,
) float64 {

	// Normalize inputs (avoid dominance)
	distanceScore := 1 / (1 + distanceKm)

	acceptanceScore := driver.AcceptanceRate
	cancellationPenalty := 1 - driver.CancellationRate
	timeoutPenalty := 1 - driver.TimeoutRate

	ratingScore := driver.Rating / 5.0

	experienceScore := math.Log(float64(driver.CompletedRides) + 1)

	// fairness boost (older idle drivers get slight boost)
	fairnessScore := time.Since(driver.LastAssignedAt).Seconds() / 3600

	// Final weighted score
	return (r.cfg.DistanceWeight * distanceScore) +
		(r.cfg.AcceptanceWeight * acceptanceScore) +
		(r.cfg.CancellationWeight * cancellationPenalty) +
		(r.cfg.TimeoutWeight * timeoutPenalty) +
		(r.cfg.RatingWeight * ratingScore) +
		(r.cfg.ExperienceWeight * experienceScore) +
		(r.cfg.FairnessWeight * fairnessScore)
}
