package ranking

import "time"

type Features struct {
	DistanceKm float64

	AcceptanceRate   float64
	CancellationRate float64
	TimeoutRate      float64

	Rating float64

	CompletedTrips int

	LastAssignedAt time.Time
}
