package expansion

import (
	"context"

	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/discovery/contract"
	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/discovery/lookup"
	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/discovery/search"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/geo"
)

type RingExpander struct {
	h3     *geo.H3Service
	lookup lookup.CellDriverLookup
	policy RingExpansionPolicy
}

func NewRingExpander(
	h3 *geo.H3Service,
	lookup lookup.CellDriverLookup,
	policy RingExpansionPolicy,
) *RingExpander {

	return &RingExpander{
		h3:     h3,
		lookup: lookup,
		policy: policy,
	}
}

func (e *RingExpander) Search(
	ctx context.Context,
	state *search.SearchState,
) error {

	for {
		decision, reason, err := e.policy.Next(ctx, state)
		if err != nil {
			return err
		}

		if reason != contract.SearchContinue {
			state.TerminationReason = reason
			return nil
		}

		cells, err := e.h3.CellsInRing(
			state.CenterCell,
			decision.Ring,
		)
		if err != nil {
			return err
		}

		state.CurrentRing = decision.NextRing

		state.VisitCells(len(cells))
		state.VisitRing()

		state.LastDensity = decision.Density

		driverIDs, err := e.lookup.GetDriversInCells(
			ctx,
			cells,
			state.Budget.RemainingCandidates,
		)
		if err != nil {
			return err
		}

		state.AddDrivers(driverIDs)
	}
}
