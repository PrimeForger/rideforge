package refinement

import (
	"context"

	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/discovery/search"
)

type NoopRefiner struct{}

func NewNoopRefiner() *NoopRefiner {
	return &NoopRefiner{}
}

func (r *NoopRefiner) Refine(
	ctx context.Context,
	result search.Result,
	req search.Request,
) (search.Result, error) {

	return result, nil
}

var _ Refiner = (*NoopRefiner)(nil)
