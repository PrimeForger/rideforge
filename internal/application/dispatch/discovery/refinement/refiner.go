package refinement

import (
	"context"

	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/discovery/search"
)

type Refiner interface {
	Refine(
		ctx context.Context,
		result search.Result,
		req search.Request,
	) (search.Result, error)
}
