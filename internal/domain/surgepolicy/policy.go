package surgepolicy

import (
	"errors"
	"math"
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusActive   Status = "ACTIVE"
	StatusInactive Status = "INACTIVE"
)

var (
	ErrInvalidPolicy       = errors.New("invalid surge policy")
	ErrOverlappingPolicies = errors.New("overlapping active surge policies for locality")
)

type Policy struct {
	ID             uuid.UUID
	LocalityID     uuid.UUID
	Version        int64
	Status         Status
	MinMultiplier  float64
	MaxMultiplier  float64
	MinimumDemand  int
	RatioTiers     []RatioTier
	EffectiveFrom  time.Time
	EffectiveUntil *time.Time
}

// RatioTier applies Multiplier when the occupancy imbalance ratio is at least
// MinImbalanceRatio. Tiers are ordered by increasing threshold.
type RatioTier struct {
	MinImbalanceRatio float64
	Multiplier        float64
}

func (p Policy) ValidAt(now time.Time) bool {
	if err := p.Validate(); err != nil || p.Status != StatusActive || now.Before(p.EffectiveFrom) {
		return false
	}
	return p.EffectiveUntil == nil || now.Before(*p.EffectiveUntil)
}

// MultiplierFor returns the highest configured tier reached by the current
// imbalance ratio, capped by this locality's policy maximum. When demand is
// below the locality minimum, the base 1.00 multiplier applies.
func (p Policy) MultiplierFor(demand int, imbalanceRatio float64, now time.Time) (float64, error) {
	if !p.ValidAt(now) || demand < 0 || math.IsNaN(imbalanceRatio) || math.IsInf(imbalanceRatio, 0) || imbalanceRatio < 0 {
		return 0, ErrInvalidPolicy
	}
	if demand < p.MinimumDemand {
		return 1.0, nil
	}
	result := p.MinMultiplier
	for _, tier := range p.RatioTiers {
		if imbalanceRatio < tier.MinImbalanceRatio {
			break
		}
		result = tier.Multiplier
	}
	if result > p.MaxMultiplier {
		result = p.MaxMultiplier
	}
	return result, nil
}

func (p Policy) Validate() error {
	if p.ID == uuid.Nil || p.LocalityID == uuid.Nil || p.Version < 1 || p.Status != StatusActive && p.Status != StatusInactive ||
		!finite(p.MinMultiplier) || !finite(p.MaxMultiplier) || p.MinMultiplier < 1 || p.MaxMultiplier < p.MinMultiplier || p.MaxMultiplier > 1.75 || p.EffectiveFrom.IsZero() || p.MinimumDemand < 0 {
		return ErrInvalidPolicy
	}
	if p.Status == StatusActive && len(p.RatioTiers) == 0 {
		return ErrInvalidPolicy
	}
	for i, tier := range p.RatioTiers {
		if !finite(tier.MinImbalanceRatio) || !finite(tier.Multiplier) || tier.MinImbalanceRatio <= 0 || tier.Multiplier < p.MinMultiplier || tier.Multiplier > p.MaxMultiplier || i > 0 && (tier.MinImbalanceRatio <= p.RatioTiers[i-1].MinImbalanceRatio || tier.Multiplier < p.RatioTiers[i-1].Multiplier) {
			return ErrInvalidPolicy
		}
	}
	if p.EffectiveUntil != nil && !p.EffectiveUntil.After(p.EffectiveFrom) {
		return ErrInvalidPolicy
	}
	return nil
}

// ValidateNoOverlappingActivePolicies ensures active policies for a locality
// have disjoint half-open effective periods. Adjacent periods are valid.
func ValidateNoOverlappingActivePolicies(policies []Policy) error {
	for i, current := range policies {
		if current.Status != StatusActive {
			continue
		}
		for _, next := range policies[i+1:] {
			if next.Status != StatusActive || current.LocalityID != next.LocalityID {
				continue
			}
			if (current.EffectiveUntil == nil || next.EffectiveFrom.Before(*current.EffectiveUntil)) &&
				(next.EffectiveUntil == nil || current.EffectiveFrom.Before(*next.EffectiveUntil)) {
				return ErrOverlappingPolicies
			}
		}
	}
	return nil
}

func finite(value float64) bool { return !math.IsNaN(value) && !math.IsInf(value, 0) }
