package search

import "github.com/ashadashraf/ride-hail-app/internal/application/dispatch/candidate"

type Result struct {
	// DriverIDs []uuid.UUID
	// Iterator   candidate.Iterator
	Candidates *candidate.Collection

	Backend string

	RadiusKm float64

	RingsVisited int

	CellsVisited int
}
