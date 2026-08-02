package stage

import (
	"context"

	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/candidate"
	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/candidate/pipeline"
)

type Ranker interface {
	Score(
		ctx context.Context,
		pipelineCtx *pipeline.Context,
		candidate *candidate.Candidate,
	) (float64, error)
}
