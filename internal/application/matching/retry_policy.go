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
	// Current retry number (0 = first attempt)
	AttemptCount int

	// Drivers returned by candidate discovery before filtering.
	NearbyDriverCount int

	// Drivers remaining after the candidate pipeline.
	EligibleDriverCount int

	// Historical offer statistics for this ride.
	PreviouslyOfferedCount  int
	PreviouslyAcceptedCount int
	PreviouslyRejectedCount int
	PreviouslyTimedOutCount int
}

func NewRetryPolicy(cfg config.MatchingRetryConfig) *RetryPolicy {
	return &RetryPolicy{
		cfg: cfg,
	}
}

func (p *RetryPolicy) Decide(
	input RetryInput,
) RetryDecision {

	return RetryDecision{
		RadiusKm: p.decideRadius(input),

		OfferBatchSize: p.decideBatchSize(input),

		CandidateLimit: p.decideCandidateLimit(input),

		OfferTimeout: p.decideOfferTimeout(input),
	}
}

func (p *RetryPolicy) decideRadius(
	input RetryInput,
) float64 {

	radius := p.cfg.BaseRadiusKm *
		math.Pow(1.6, float64(input.AttemptCount))

	// Very low supply -> expand slightly faster.
	if input.EligibleDriverCount > 0 &&
		input.EligibleDriverCount <= 3 {

		radius *= 1.15
	}

	// Plenty of candidates already.
	// Don't keep expanding aggressively.
	if input.EligibleDriverCount >=
		p.cfg.BaseCandidateLimit {

		radius *= 0.90
	}

	return clampFloat(
		radius,
		p.cfg.BaseRadiusKm,
		p.cfg.MaxRadiusKm,
	)
}

func (p *RetryPolicy) decideBatchSize(
	input RetryInput,
) int {

	batch := p.cfg.BaseOfferBatchSize

	// Increase batch size on later retries.
	switch {
	case input.AttemptCount >= 2:
		batch += 2
	case input.AttemptCount >= 1:
		batch += 1
	}

	// If very few eligible drivers remain,
	// offer to more of them immediately.
	if input.EligibleDriverCount > 0 &&
		input.EligibleDriverCount <= 3 {

		batch++
	}

	// If we already offered many drivers for this ride,
	// avoid growing the batch forever.
	if input.PreviouslyOfferedCount >= 10 {
		batch--
	}

	// If almost everyone previously timed out,
	// increase parallelism slightly.
	if input.PreviouslyTimedOutCount >= 3 {
		batch++
	}

	// Never exceed the available candidate pool.
	if input.EligibleDriverCount > 0 &&
		batch > input.EligibleDriverCount {

		batch = input.EligibleDriverCount
	}

	return clampInt(
		batch,
		1,
		p.cfg.MaxOfferBatchSize,
	)
}

func (p *RetryPolicy) decideCandidateLimit(
	input RetryInput,
) int {

	limit := p.cfg.BaseCandidateLimit

	// Later retries search more candidates.
	limit += input.AttemptCount * 25

	// Low supply -> widen search.
	if input.EligibleDriverCount > 0 &&
		input.EligibleDriverCount <= 5 {

		limit += 20
	}

	// Already found plenty.
	// Don't ask discovery for unnecessary work.
	if input.NearbyDriverCount >=
		p.cfg.BaseCandidateLimit {

		limit -= 10
	}

	return clampInt(
		limit,
		p.cfg.BaseCandidateLimit,
		p.cfg.MaxCandidateLimit,
	)
}

func (p *RetryPolicy) decideOfferTimeout(
	input RetryInput,
) time.Duration {

	timeoutMs := p.cfg.BaseOfferTimeoutMs

	// Later retries shouldn't wait as long.
	switch {
	case input.AttemptCount >= 2:
		timeoutMs -= 2000
	case input.AttemptCount >= 1:
		timeoutMs -= 1000
	}

	// Small driver supply → give drivers
	// more time to respond.
	if input.EligibleDriverCount > 0 &&
		input.EligibleDriverCount <= 3 {

		timeoutMs += 2000
	}

	// If previous offers frequently timed out,
	// extend timeout slightly.
	if input.PreviouslyTimedOutCount >= 3 {
		timeoutMs += 1000
	}

	// If we already offered many drivers,
	// don't keep everyone waiting.
	if input.PreviouslyOfferedCount >= 10 {
		timeoutMs -= 1000
	}

	timeoutMs = clampInt(
		timeoutMs,
		p.cfg.MinOfferTimeoutMs,
		p.cfg.MaxOfferTimeoutMs,
	)

	return time.Duration(timeoutMs) * time.Millisecond
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
