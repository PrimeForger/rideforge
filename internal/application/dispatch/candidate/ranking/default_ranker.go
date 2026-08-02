package ranking

import (
	"context"

	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/candidate"
	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/candidate/pipeline"
)

type DefaultRanker struct {
	extractor FeatureExtractor
	scorer    Scorer
}

func NewDefaultRanker(
	extractor FeatureExtractor,
	scorer Scorer,
) *DefaultRanker {

	return &DefaultRanker{
		extractor: extractor,
		scorer:    scorer,
	}
}

func (r *DefaultRanker) Rank(
	ctx context.Context,
	pipelineCtx *pipeline.Context,
	driverCandidate *candidate.Candidate,
) (candidate.Score, error) {

	features, err := r.extractor.Extract(
		ctx,
		pipelineCtx,
		driverCandidate,
	)
	if err != nil {
		return candidate.Score{}, err
	}

	return r.scorer.Score(
		ctx,
		features,
	)
}
