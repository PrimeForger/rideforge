package occupancy_test

import (
	"testing"
	"time"

	"github.com/ashadashraf/ride-hail-app/internal/domain/occupancy"
)

func TestNewCellOccupancy(t *testing.T) {
	now := time.Now()
	warmThreshold := 1.2
	hotThreshold := 1.5
	minDemandForHot := 2

	tests := []struct {
		name           string
		cellID         string
		supply         int
		demand         int
		expectedRatio  float64
		expectedStatus occupancy.CellStatus
		isHot          bool
		isWarm         bool
		isNormal       bool
	}{
		{
			name:           "Zero supply, zero demand",
			cellID:         "8860145261fffff",
			supply:         0,
			demand:         0,
			expectedRatio:  0.0,
			expectedStatus: occupancy.StatusNormal,
			isNormal:       true,
		},
		{
			name:           "Zero supply, 1 demand (< minDemandForHot)",
			cellID:         "8860145261fffff",
			supply:         0,
			demand:         1,
			expectedRatio:  1.0,
			expectedStatus: occupancy.StatusNormal, // ratio 1.0 < 1.2
			isNormal:       true,
		},
		{
			name:           "Zero supply, 2 demand (>= minDemandForHot and ratio >= 1.5)",
			cellID:         "8860145261fffff",
			supply:         0,
			demand:         2,
			expectedRatio:  2.0,
			expectedStatus: occupancy.StatusHot,
			isHot:          true,
		},
		{
			name:           "Balanced supply and demand",
			cellID:         "8860145261fffff",
			supply:         10,
			demand:         10,
			expectedRatio:  1.0,
			expectedStatus: occupancy.StatusNormal,
			isNormal:       true,
		},
		{
			name:           "High supply, low demand",
			cellID:         "8860145261fffff",
			supply:         20,
			demand:         5,
			expectedRatio:  0.25,
			expectedStatus: occupancy.StatusNormal,
			isNormal:       true,
		},
		{
			name:           "Moderate imbalance (Warm status)",
			cellID:         "8860145261fffff",
			supply:         10,
			demand:         13,
			expectedRatio:  1.3,
			expectedStatus: occupancy.StatusWarm,
			isWarm:         true,
		},
		{
			name:           "Critical imbalance (Hot status)",
			cellID:         "8860145261fffff",
			supply:         10,
			demand:         20,
			expectedRatio:  2.0,
			expectedStatus: occupancy.StatusHot,
			isHot:          true,
		},
		{
			name:           "Negative supply and demand sanitized",
			cellID:         "8860145261fffff",
			supply:         -5,
			demand:         -3,
			expectedRatio:  0.0,
			expectedStatus: occupancy.StatusNormal,
			isNormal:       true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			occ := occupancy.NewCellOccupancy(
				tc.cellID,
				tc.supply,
				tc.demand,
				warmThreshold,
				hotThreshold,
				minDemandForHot,
				now,
			)

			if occ.CellID != tc.cellID {
				t.Errorf("expected cellID %s, got %s", tc.cellID, occ.CellID)
			}
			if occ.SupplyCount < 0 {
				t.Errorf("supply count should be non-negative, got %d", occ.SupplyCount)
			}
			if occ.DemandCount < 0 {
				t.Errorf("demand count should be non-negative, got %d", occ.DemandCount)
			}
			if occ.ImbalanceRatio != tc.expectedRatio {
				t.Errorf("expected ratio %f, got %f", tc.expectedRatio, occ.ImbalanceRatio)
			}
			if occ.Status != tc.expectedStatus {
				t.Errorf("expected status %s, got %s", tc.expectedStatus, occ.Status)
			}
			if occ.IsHot() != tc.isHot {
				t.Errorf("expected IsHot() == %v, got %v", tc.isHot, occ.IsHot())
			}
			if occ.IsWarm() != tc.isWarm {
				t.Errorf("expected IsWarm() == %v, got %v", tc.isWarm, occ.IsWarm())
			}
			if occ.IsNormal() != tc.isNormal {
				t.Errorf("expected IsNormal() == %v, got %v", tc.isNormal, occ.IsNormal())
			}
		})
	}
}
