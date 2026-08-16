package strategy

import (
	"context"

	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/candidate"
	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/discovery/expansion"
	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/discovery/refinement"
	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/discovery/search"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/geo"
)

type H3Strategy struct {
	h3            *geo.H3Service
	expander      *expansion.RingExpander
	budgetFactory search.BudgetFactory
	refiner       refinement.Refiner
}

func NewH3Strategy(
	h3 *geo.H3Service,
	expander *expansion.RingExpander,
	budgetFactory search.BudgetFactory,
	refiner refinement.Refiner,
) *H3Strategy {
	return &H3Strategy{
		h3:            h3,
		expander:      expander,
		budgetFactory: budgetFactory,
		refiner:       refiner,
	}
}

func (s *H3Strategy) FindCandidates(
	ctx context.Context,
	req search.Request,
) (search.Result, error) {

	centerCell, err := s.h3.CenterCell(
		req.PickupLat,
		req.PickupLng,
	)
	if err != nil {
		return search.Result{}, err
	}

	budget := s.budgetFactory.NewBudget(
		search.PolicyInput{
			MatchingAttempt: req.MatchingAttempt,
			CandidateLimit:  req.CandidateLimit,
		},
	)

	state := search.NewSearchState(
		centerCell,
		budget,
	)

	if err := s.expander.Search(
		ctx,
		state,
	); err != nil {
		return search.Result{}, err
	}

	collection := candidate.NewCollectionFromIDs(
		state.DriverIDs,
	)

	result := search.Result{
		Candidates: collection,

		Backend: "h3",

		RadiusKm: req.RadiusKm,

		CellsVisited: state.CellsVisited,

		RingsVisited: state.RingsVisited,
	}

	return s.refiner.Refine(
		ctx,
		result,
		search.Request{
			PickupLat:       req.PickupLat,
			PickupLng:       req.PickupLng,
			RadiusKm:        req.RadiusKm,
			CandidateLimit:  req.CandidateLimit,
			MatchingAttempt: req.MatchingAttempt,
		},
	)
}
