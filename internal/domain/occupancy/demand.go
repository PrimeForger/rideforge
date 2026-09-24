package occupancy

import (
	"time"

	"github.com/google/uuid"
)

type CellDemand struct {
	CellID      string    `json:"cell_id"`
	ActiveCount int       `json:"active_count"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type RideDemandPointer struct {
	RideID    uuid.UUID `json:"ride_id"`
	CellID    string    `json:"cell_id"`
	CreatedAt time.Time `json:"created_at"`
}
