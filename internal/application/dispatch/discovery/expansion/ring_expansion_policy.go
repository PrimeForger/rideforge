package expansion

import (
	"context"

	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/discovery/contract"
	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/discovery/search"
)

type RingExpansionPolicy interface {
	Next(
		ctx context.Context,
		state *search.SearchState,
	) (RingDecision, contract.SearchTerminationReason, error)
}
