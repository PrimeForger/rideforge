package refinement

import "github.com/google/uuid"

type Request struct {
	PickupLat float64
	PickupLng float64

	CandidateIDs []uuid.UUID

	RadiusKm float64
}
