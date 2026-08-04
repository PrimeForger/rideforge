package stage

import (
	"context"

	"github.com/google/uuid"

	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/candidate"
	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/candidate/pipeline"
)

type DuplicateFilter struct{}

func NewDuplicateFilter() *DuplicateFilter {
	return &DuplicateFilter{}
}

func (f *DuplicateFilter) Execute(
	ctx context.Context,
	pipelineCtx *pipeline.Context,
	candidates *candidate.Collection,
) error {

	before := candidates.Len()

	seen := make(map[uuid.UUID]struct{}, before)

	candidates.RemoveIf(func(c *candidate.Candidate) bool {

		if _, exists := seen[c.ID]; exists {
			return true
		}

		seen[c.ID] = struct{}{}

		return false
	})

	pipelineCtx.Result.FilteredCandidates += before - candidates.Len()

	return nil
}

var _ pipeline.Stage = (*DuplicateFilter)(nil)
