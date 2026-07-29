package profile

import "github.com/ashadashraf/ride-hail-app/internal/application/dispatch/discovery/search"

type DefaultExpansionPolicy struct {
	profile ExpansionProfile
}

func NewDefaultExpansionPolicy(
	profile ExpansionProfile,
) *DefaultExpansionPolicy {

	return &DefaultExpansionPolicy{
		profile: profile,
	}
}

func (p *DefaultExpansionPolicy) ConfigureBudget(
	budget *search.SearchBudget,
) {

	budget.ConfigureLimits(
		p.profile.MaxRing,
		p.profile.MaxCells,
		p.profile.MaxCandidates,
	)
}

func (p *DefaultExpansionPolicy) NextRing(
	currentRing int,
) int {

	next := currentRing + p.profile.RingStep

	if next > p.profile.MaxRing {
		return p.profile.MaxRing
	}

	return next
}
