package matching

import (
	"math"
	"time"

	"github.com/ashadashraf/ride-hail-app/internal/config"
)

// Determine dispatch parameters for the current matching attempt.
//
// Inputs:
//   - Retry count
//   - Candidate availability
//
// Outputs:
//   - Search radius
//   - Candidate limit
//   - Offer batch size
//   - Offer timeout

type RetryPolicy struct {
	cfg config.MatchingRetryConfig
}

type RetryDecision struct {
	RadiusKm       float64
	OfferBatchSize int
	CandidateLimit int
	OfferTimeout   time.Duration
}

type RetryInput struct {
	AttemptCount        int
	NearbyDriverCount   int
	EligibleDriverCount int
}

func NewRetryPolicy(cfg config.MatchingRetryConfig) *RetryPolicy {
	return &RetryPolicy{
		cfg: cfg,
	}
}

func (p *RetryPolicy) Decide(input RetryInput) RetryDecision {
	attempt := input.AttemptCount

	radius := p.cfg.BaseRadiusKm * math.Pow(1.6, float64(attempt))
	radius = clampFloat(radius, p.cfg.BaseRadiusKm, p.cfg.MaxRadiusKm)

	batchSize := p.cfg.BaseOfferBatchSize

	if attempt >= 1 {
		batchSize++
	}

	if attempt >= 2 {
		batchSize++
	}

	if input.EligibleDriverCount > 0 && input.EligibleDriverCount <= 3 {
		batchSize = minInt(batchSize+1, p.cfg.MaxOfferBatchSize)
	}

	batchSize = clampInt(batchSize, 1, p.cfg.MaxOfferBatchSize)

	candidateLimit := p.cfg.BaseCandidateLimit + attempt*25
	candidateLimit = clampInt(candidateLimit, p.cfg.BaseCandidateLimit, p.cfg.MaxCandidateLimit)

	timeoutMs := p.cfg.BaseOfferTimeoutMs

	// More retries means we reduce waiting time slightly.
	if attempt >= 1 {
		timeoutMs -= 1000
	}

	if attempt >= 2 {
		timeoutMs -= 1000
	}

	// If driver supply is low, give drivers a bit more time.
	if input.EligibleDriverCount > 0 && input.EligibleDriverCount <= 3 {
		timeoutMs += 2000
	}

	timeoutMs = clampInt(timeoutMs, p.cfg.MinOfferTimeoutMs, p.cfg.MaxOfferTimeoutMs)

	return RetryDecision{
		RadiusKm:       radius,
		OfferBatchSize: batchSize,
		CandidateLimit: candidateLimit,
		OfferTimeout:   time.Duration(timeoutMs) * time.Millisecond,
	}
}

func clampFloat(v, min, max float64) float64 {
	if v < min {
		return min
	}

	if v > max {
		return max
	}

	return v
}

func clampInt(v, min, max int) int {
	if v < min {
		return min
	}

	if v > max {
		return max
	}

	return v
}

func minInt(a, b int) int {
	if a < b {
		return a
	}

	return b
}
