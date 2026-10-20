package surgepolicy

import (
	"github.com/google/uuid"
	"math"
	"testing"
	"time"
)

func TestPolicyValidity(t *testing.T) {
	now := time.Now()
	p := Policy{ID: uuid.New(), LocalityID: uuid.New(), Version: 1, Status: StatusActive, MinMultiplier: 1, MaxMultiplier: 1.5, EffectiveFrom: now.Add(-time.Minute), RatioTiers: []RatioTier{{MinImbalanceRatio: 1, Multiplier: 1.1}}}
	if err := p.Validate(); err != nil || !p.ValidAt(now) {
		t.Fatalf("valid policy rejected: %v", err)
	}
	p.MaxMultiplier = .9
	if err := p.Validate(); err == nil {
		t.Fatal("invalid multiplier range accepted")
	}
}

func TestPolicyRejectsNonFiniteMultipliersAndThresholds(t *testing.T) {
	now := time.Now()
	base := Policy{ID: uuid.New(), LocalityID: uuid.New(), Version: 1, Status: StatusActive, MinMultiplier: 1, MaxMultiplier: 1.5, EffectiveFrom: now.Add(-time.Minute)}
	for _, mutate := range []func(*Policy){
		func(p *Policy) { p.MinMultiplier = math.NaN() },
		func(p *Policy) { p.MaxMultiplier = math.Inf(1) },
		func(p *Policy) { p.RatioTiers = []RatioTier{{MinImbalanceRatio: math.NaN(), Multiplier: 1.2}} },
	} {
		p := base
		mutate(&p)
		if err := p.Validate(); err == nil {
			t.Fatal("accepted non-finite policy value")
		}
	}
}

func TestActivePolicyRequiresValidMonotonicTiers(t *testing.T) {
	now := time.Now()
	base := Policy{
		ID: uuid.New(), LocalityID: uuid.New(), Version: 1, Status: StatusActive,
		MinMultiplier: 1, MaxMultiplier: 1.5, EffectiveFrom: now.Add(-time.Minute),
		RatioTiers: []RatioTier{{MinImbalanceRatio: 1, Multiplier: 1.1}, {MinImbalanceRatio: 2, Multiplier: 1.5}},
	}
	cases := []struct {
		name   string
		mutate func(*Policy)
	}{
		{name: "no tiers", mutate: func(p *Policy) { p.RatioTiers = nil }},
		{name: "unordered thresholds", mutate: func(p *Policy) { p.RatioTiers[1].MinImbalanceRatio = 0.5 }},
		{name: "decreasing multipliers", mutate: func(p *Policy) { p.RatioTiers[1].Multiplier = 1.05 }},
		{name: "tier above policy max", mutate: func(p *Policy) { p.RatioTiers[1].Multiplier = 1.75 }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := base
			p.RatioTiers = append([]RatioTier(nil), base.RatioTiers...)
			tc.mutate(&p)
			if err := p.Validate(); err == nil {
				t.Fatal("invalid active tier configuration accepted")
			}
		})
	}
	if err := base.Validate(); err != nil {
		t.Fatalf("valid tiers rejected: %v", err)
	}
}

func TestValidateNoOverlappingActivePolicySchedules(t *testing.T) {
	now := time.Now()
	localityID := uuid.New()
	policy := func(version int64, from time.Time, until *time.Time) Policy {
		return Policy{
			ID: uuid.New(), LocalityID: localityID, Version: version, Status: StatusActive,
			MinMultiplier: 1, MaxMultiplier: 1.5, EffectiveFrom: from, EffectiveUntil: until,
			RatioTiers: []RatioTier{{MinImbalanceRatio: 1, Multiplier: 1.1}},
		}
	}
	currentEnd := now.Add(time.Hour)
	cases := []struct {
		name     string
		policies []Policy
		wantErr  bool
	}{
		{name: "one current", policies: []Policy{policy(1, now.Add(-time.Hour), &currentEnd)}},
		{name: "historical and current", policies: []Policy{policy(1, now.Add(-3*time.Hour), ptrTime(now.Add(-time.Hour))), policy(2, now.Add(-time.Hour), &currentEnd)}},
		{name: "current and future", policies: []Policy{policy(1, now.Add(-time.Hour), &currentEnd), policy(2, currentEnd, nil)}},
		{name: "overlapping", policies: []Policy{policy(1, now.Add(-time.Hour), &currentEnd), policy(2, now, nil)}, wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateNoOverlappingActivePolicies(tc.policies)
			if (err != nil) != tc.wantErr {
				t.Fatalf("ValidateNoOverlappingActivePolicies() error = %v, want error %t", err, tc.wantErr)
			}
		})
	}
}

func ptrTime(value time.Time) *time.Time { return &value }

func TestPolicyMultiplierSelectsOrderedTierAndCapsAtConfiguredMaximum(t *testing.T) {
	p := Policy{
		ID: uuid.New(), LocalityID: uuid.New(), Version: 1, Status: StatusActive,
		MinMultiplier: 1, MaxMultiplier: 1.5, MinimumDemand: 3,
		EffectiveFrom: time.Now().Add(-time.Minute),
		RatioTiers: []RatioTier{
			{MinImbalanceRatio: 1, Multiplier: 1.1},
			{MinImbalanceRatio: 1.2, Multiplier: 1.2},
			{MinImbalanceRatio: 1.5, Multiplier: 1.35},
			{MinImbalanceRatio: 2, Multiplier: 1.5},
			{MinImbalanceRatio: 3, Multiplier: 1.75},
		},
	}
	if err := p.Validate(); err == nil {
		t.Fatal("tier above locality maximum was accepted")
	}
	p.MaxMultiplier = 1.75
	if err := p.Validate(); err != nil {
		t.Fatalf("explicit 1.75 maximum rejected: %v", err)
	}
	p.RatioTiers[2].MinImbalanceRatio = 1.1
	if err := p.Validate(); err == nil {
		t.Fatal("unordered ratio tiers were accepted")
	}
}

func TestPolicyMultiplierForMinimumDemandAndTierBoundaries(t *testing.T) {
	now := time.Now()
	p := Policy{
		ID: uuid.New(), LocalityID: uuid.New(), Version: 1, Status: StatusActive,
		MinMultiplier: 1, MaxMultiplier: 1.5, MinimumDemand: 3,
		EffectiveFrom: now.Add(-time.Minute),
		RatioTiers: []RatioTier{
			{MinImbalanceRatio: 1, Multiplier: 1.1},
			{MinImbalanceRatio: 1.2, Multiplier: 1.2},
			{MinImbalanceRatio: 1.5, Multiplier: 1.35},
			{MinImbalanceRatio: 2, Multiplier: 1.5},
		},
	}
	for _, tc := range []struct {
		demand int
		ratio  float64
		want   float64
	}{
		{demand: 2, ratio: 5, want: 1},
		{demand: 3, ratio: .99, want: 1},
		{demand: 3, ratio: 1, want: 1.1},
		{demand: 3, ratio: 1.2, want: 1.2},
		{demand: 8, ratio: 2, want: 1.5},
	} {
		got, err := p.MultiplierFor(tc.demand, tc.ratio, now)
		if err != nil || got != tc.want {
			t.Errorf("MultiplierFor(%d, %.2f) = %.2f, %v; want %.2f", tc.demand, tc.ratio, got, err, tc.want)
		}
	}
}
