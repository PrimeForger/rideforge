package strategy

import (
	"context"

	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/discovery/search"
)

// Discover nearby candidate drivers.
// This component performs spatial search only.
// It does not load driver data or determine dispatch order.

type CandidateSearcher interface {
	FindCandidates(ctx context.Context, req search.Request) (search.Result, error)
}
