package ranking

import (
	"context"

	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/candidate"
	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/candidate/pipeline"
)

type Ranker interface {
	Rank(
		ctx context.Context,
		pipelineCtx *pipeline.Context,
		candidate *candidate.Candidate,
	) (candidate.Score, error)
}
