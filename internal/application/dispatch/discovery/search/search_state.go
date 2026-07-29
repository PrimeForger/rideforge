package search

import (
	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/discovery/contract"
	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/discovery/density"
	"github.com/google/uuid"
)

type SearchState struct {
	CenterCell string

	Budget SearchBudget

	DriverIDs []uuid.UUID
	Seen      map[uuid.UUID]struct{}

	CurrentRing int

	RingsVisited int
	CellsVisited int

	LastDensity       density.DensityClass
	TerminationReason contract.SearchTerminationReason
}

func NewSearchState(
	centerCell string,
	budget SearchBudget,
) *SearchState {

	return &SearchState{
		CenterCell: centerCell,

		Budget: budget,

		DriverIDs: make([]uuid.UUID, 0, budget.CandidateLimit),
		Seen:      make(map[uuid.UUID]struct{}),

		CurrentRing: 0,
	}
}

func (s *SearchState) AddDrivers(
	driverIDs []uuid.UUID,
) {

	added := 0

	for _, id := range driverIDs {

		if _, exists := s.Seen[id]; exists {
			continue
		}

		s.Seen[id] = struct{}{}
		s.DriverIDs = append(s.DriverIDs, id)

		added++
	}

	s.Budget.ConsumeCandidates(added)
}

func (s *SearchState) VisitRing() {
	s.RingsVisited++
	s.Budget.ConsumeRing()
}

func (s *SearchState) VisitCells(count int) {
	s.CellsVisited += count
	s.Budget.ConsumeCells(count)
}

func (s *SearchState) IsSatisfied() bool {

	return s.Budget.CandidateBudgetExhausted() ||
		s.Budget.RingBudgetExhausted() ||
		s.Budget.CellBudgetExhausted()
}
