package selector

import (
	"context"

	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/discovery/profile"
	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/discovery/search"
)

type Selector interface {
	Select(
		ctx context.Context,
		state *search.SearchState,
	) (profile.SearchProfile, error)
}
