package search

type Request struct {
	PickupLat float64
	PickupLng float64

	RadiusKm float64

	CandidateLimit int

	MatchingAttempt int
}
