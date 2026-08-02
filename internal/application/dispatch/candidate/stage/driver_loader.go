package stage

import (
	"context"

	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/candidate"
)

// Load driver information for every discovered candidate in a single batch.
// Subsequent stages operate only on enriched candidates.

type DriverLoader interface {
	Load(
		ctx context.Context,
		candidates *candidate.Collection,
	) error
}
