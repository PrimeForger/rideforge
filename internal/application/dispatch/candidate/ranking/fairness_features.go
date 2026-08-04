package ranking

import "time"

type FairnessFeatures struct {
	LastAssignedAt time.Time
	// IdleDuration       time.Time
	// RecentAssignments  int32
	// ConsecutiveRejects int32
}
