package pipeline

import (
	"context"

	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/discovery/profile"
	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/discovery/search"
)

type Pipeline interface {
	Execute(
		ctx context.Context,
		state *search.SearchState,
		builder *profile.Builder,
	) error
}
