package expansion

import (
	"context"

	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/discovery/contract"
	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/discovery/search"
	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/discovery/selector"
)

type AdaptiveRingExpansionPolicy struct {
	selector selector.Selector
}

func NewAdaptiveRingExpansionPolicy(
	selector selector.Selector,
) *AdaptiveRingExpansionPolicy {

	return &AdaptiveRingExpansionPolicy{
		selector: selector,
	}
}

func (p *AdaptiveRingExpansionPolicy) Next(
	ctx context.Context,
	state *search.SearchState,
) (RingDecision, contract.SearchTerminationReason, error) {

	// Candidate budget exhausted.
	if state.Budget.CandidateBudgetExhausted() {
		return RingDecision{}, contract.SearchCandidateBudgetExhausted, nil
	}

	// Ring budget exhausted.
	if state.Budget.RingBudgetExhausted() {
		return RingDecision{}, contract.SearchRingBudgetExhausted, nil
	}

	// Cell budget exhausted.
	if state.Budget.CellBudgetExhausted() {
		return RingDecision{}, contract.SearchCellBudgetExhausted, nil
	}

	// Budget exhausted.
	// if state.IsSatisfied() {
	// 	return RingDecision{}, false, nil
	// }

	profile, err := p.selector.Select(
		ctx,
		state,
	)
	if err != nil {
		return RingDecision{}, contract.SearchContinue, err
	}

	expansionPolicy := profile.Expansion()

	expansionPolicy.ConfigureBudget(
		&state.Budget,
	)

	nextRing := expansionPolicy.NextRing(
		state.CurrentRing,
	)

	return RingDecision{
		Ring:     state.CurrentRing,
		NextRing: nextRing,
		Density:  state.LastDensity,
	}, contract.SearchContinue, nil
}
