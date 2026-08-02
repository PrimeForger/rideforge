package stage

import (
	"context"

	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/candidate"
	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/candidate/pipeline"
)

// Remove drivers that have already received an offer for this ride.

type AlreadyOfferedFilter struct {
	offeredDrivers OfferedDriverProvider
}

func NewAlreadyOfferedFilter(
	offeredDrivers OfferedDriverProvider,
) *AlreadyOfferedFilter {

	return &AlreadyOfferedFilter{
		offeredDrivers: offeredDrivers,
	}
}

func (f *AlreadyOfferedFilter) Execute(
	ctx context.Context,
	pipelineCtx *pipeline.Context,
	candidates *candidate.Collection,
) error {

	before := candidates.Len()

	offeredSet, err := f.offeredDrivers.GetOfferedDrivers(
		ctx,
		pipelineCtx.RideID,
	)
	if err != nil {
		return err
	}

	candidates.RemoveIf(func(c *candidate.Candidate) bool {

		if c.Driver == nil {
			return true
		}

		_, exists := offeredSet[c.Driver.ID]
		return exists
	})

	pipelineCtx.Result.FilteredCandidates += before - candidates.Len()

	return nil
}

var _ pipeline.Stage = (*AlreadyOfferedFilter)(nil)
