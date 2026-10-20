package ports

import (
	"context"
	"github.com/ashadashraf/ride-hail-app/internal/domain/surgepolicy"
	"time"
)

type SurgePolicyRepository interface {
	LoadActive(ctx context.Context, now time.Time) ([]surgepolicy.Policy, error)
}

type SurgePolicyProvider interface {
	PolicyForLocality(localityID string) (surgepolicy.Policy, bool)
	SnapshotState() (entries int, lastSuccess time.Time, lastFailure time.Time)
}

// SurgeObserver records low-cardinality quote fallback and policy refresh outcomes.
type SurgeObserver interface {
	ObserveSurgeDecision(result string)
	ObserveSurgePolicyRefresh(result string)
}
