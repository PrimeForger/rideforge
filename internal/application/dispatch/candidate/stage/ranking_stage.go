package stage

import (
	"context"

	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/candidate"
	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/candidate/pipeline"
	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/candidate/ranking"
)

// Compute ranking features and assign a score to every remaining candidate.
// Ranking is independent from dispatch orchestration.

type RankingStage struct {
	ranker ranking.Ranker
}

func NewRankingStage(
	ranker ranking.Ranker,
) *RankingStage {

	return &RankingStage{
		ranker: ranker,
	}
}

func (s *RankingStage) Execute(
	ctx context.Context,
	pipelineCtx *pipeline.Context,
	candidates *candidate.Collection,
) error {

	it := candidates.Iterator()

	for {

		candidate, ok := it.Next()
		if !ok {
			break
		}

		score, err := s.ranker.Rank(
			ctx,
			pipelineCtx,
			candidate,
		)
		if err != nil {
			return err
		}

		candidate.Metadata.Score = &score

		pipelineCtx.Result.RankedCandidates++
	}

	return nil
}

var _ pipeline.Stage = (*RankingStage)(nil)
