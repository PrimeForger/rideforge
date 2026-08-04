package stage

import (
	"context"

	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/candidate"
	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/candidate/pipeline"
	"github.com/google/uuid"
)

type ValidationFilter struct{}

func NewValidationFilter() *ValidationFilter {
	return &ValidationFilter{}
}

func (f *ValidationFilter) Execute(
	ctx context.Context,
	pipelineCtx *pipeline.Context,
	candidates *candidate.Collection,
) error {

	before := candidates.Len()

	candidates.RemoveIf(func(c *candidate.Candidate) bool {

		if c == nil {
			return true
		}

		if c.ID == uuid.Nil {
			return true
		}

		if c.Driver == nil {
			return true
		}

		return false
	})

	pipelineCtx.Result.FilteredCandidates += before - candidates.Len()

	return nil
}

var _ pipeline.Stage = (*ValidationFilter)(nil)
