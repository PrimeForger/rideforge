package ranking

import (
	"context"

	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/candidate"
	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/candidate/pipeline"
)

type Enricher interface {
	Enrich(
		ctx context.Context,
		pipelineCtx *pipeline.Context,
		candidate *candidate.Candidate,
		features *Features,
	) error
}
