package stage

import (
	"context"

	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/candidate"
	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/candidate/pipeline"
)

// Remove drivers that are not currently eligible to receive ride offers.

type AvailabilityFilter struct{}

func NewAvailabilityFilter() *AvailabilityFilter {
	return &AvailabilityFilter{}
}

func (f *AvailabilityFilter) Execute(
	ctx context.Context,
	pipelineCtx *pipeline.Context,
	candidates *candidate.Collection,
) error {

	before := candidates.Len()

	candidates.RemoveIf(func(c *candidate.Candidate) bool {

		if c.Driver == nil {
			return true
		}

		return !c.Driver.IsAvailable()
	})

	pipelineCtx.Result.FilteredCandidates += before - candidates.Len()

	return nil
}

var _ pipeline.Stage = (*AvailabilityFilter)(nil)
