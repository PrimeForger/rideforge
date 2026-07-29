package search

import "github.com/google/uuid"

type Result struct {
	DriverIDs []uuid.UUID

	Backend string

	RadiusKm float64

	RingsVisited int

	CellsVisited int
}
