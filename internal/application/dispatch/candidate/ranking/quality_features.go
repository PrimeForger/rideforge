package ranking

type QualityFeatures struct {
	AcceptanceRate   float64
	CancellationRate float64
	TimeoutRate      float64

	Rating float64

	CompletedTrips int

	// Score float64
}
