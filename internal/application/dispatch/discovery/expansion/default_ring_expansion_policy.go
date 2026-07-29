package expansion

import (
	"context"

	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/discovery/contract"
	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/discovery/search"
)

type DefaultRingExpansionPolicy struct {
	maxRing int
}

func NewDefaultRingExpansionPolicy(
	maxRing int,
) *DefaultRingExpansionPolicy {

	return &DefaultRingExpansionPolicy{
		maxRing: maxRing,
	}
}

func (p *DefaultRingExpansionPolicy) Next(
	ctx context.Context,
	state *search.SearchState,
) (RingDecision, contract.SearchTerminationReason, error) {

	if state.Budget.CandidateBudgetExhausted() {
		return RingDecision{}, contract.SearchCandidateBudgetExhausted, nil
	}

	if state.Budget.RingBudgetExhausted() {
		return RingDecision{}, contract.SearchRingBudgetExhausted, nil
	}

	if state.Budget.CellBudgetExhausted() {
		return RingDecision{}, contract.SearchCellBudgetExhausted, nil
	}

	if state.CurrentRing > p.maxRing {
		return RingDecision{}, contract.SearchRingBudgetExhausted, nil
	}

	return RingDecision{
		Ring:     state.CurrentRing,
		NextRing: state.CurrentRing + 1,
	}, contract.SearchContinue, nil
}
