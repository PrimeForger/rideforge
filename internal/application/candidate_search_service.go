package application

type CandidateSearchBackend int

const (
	CandidateSearchBackendAuto CandidateSearchBackend = iota
	CandidateSearchBackendH3
	CandidateSearchBackendGeo
)

type CandidateSearchRequest struct {
	PickupLat float64
	PickupLng float64

	CandidateLimit int

	SearchRadiusKm float64

	PreferredBackend CandidateSearchBackend
}
