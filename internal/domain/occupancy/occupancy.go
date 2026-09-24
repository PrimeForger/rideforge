package occupancy

import (
	"math"
	"time"
)

type CellStatus string

const (
	StatusNormal CellStatus = "NORMAL"
	StatusWarm   CellStatus = "WARM"
	StatusHot    CellStatus = "HOT"
)

// CellOccupancy captures the combined supply, demand, and imbalance status for an H3 cell.
type CellOccupancy struct {
	CellID         string     `json:"cell_id"`
	SupplyCount    int        `json:"supply_count"`
	DemandCount    int        `json:"demand_count"`
	ImbalanceRatio float64    `json:"imbalance_ratio"`
	Status         CellStatus `json:"status"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// NewCellOccupancy calculates the imbalance ratio and classifies the status of an H3 cell.
func NewCellOccupancy(
	cellID string,
	supplyCount int,
	demandCount int,
	warmThreshold float64,
	hotThreshold float64,
	minDemandForHot int,
	updatedAt time.Time,
) CellOccupancy {
	if supplyCount < 0 {
		supplyCount = 0
	}
	if demandCount < 0 {
		demandCount = 0
	}

	if warmThreshold <= 0 {
		warmThreshold = 1.2
	}
	if hotThreshold <= 0 {
		hotThreshold = 1.5
	}
	if minDemandForHot <= 0 {
		minDemandForHot = 2
	}

	if updatedAt.IsZero() {
		updatedAt = time.Now()
	}

	var ratio float64
	if demandCount == 0 {
		ratio = 0.0
	} else if supplyCount == 0 {
		// When supply is 0 and demand > 0, ratio represents pure unmatched demand
		ratio = float64(demandCount)
	} else {
		ratio = float64(demandCount) / float64(supplyCount)
		// Round to 4 decimal places for clean representation
		ratio = math.Round(ratio*10000) / 10000
	}

	status := StatusNormal
	if demandCount >= minDemandForHot && ratio >= hotThreshold {
		status = StatusHot
	} else if demandCount > 0 && ratio >= warmThreshold {
		status = StatusWarm
	}

	return CellOccupancy{
		CellID:         cellID,
		SupplyCount:    supplyCount,
		DemandCount:    demandCount,
		ImbalanceRatio: ratio,
		Status:         status,
		UpdatedAt:      updatedAt,
	}
}

func (o CellOccupancy) IsHot() bool {
	return o.Status == StatusHot
}

func (o CellOccupancy) IsWarm() bool {
	return o.Status == StatusWarm
}

func (o CellOccupancy) IsNormal() bool {
	return o.Status == StatusNormal
}
