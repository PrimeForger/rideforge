package builders

import (
	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/candidate/ranking"
	"github.com/ashadashraf/ride-hail-app/internal/config"
)

func BuildRanker(
	cfg *config.RankingConfig,
) ranking.Ranker {

	featureExtractor := ranking.NewDefaultFeatureExtractor()

	scorer := ranking.NewDefaultScorer(cfg)

	return ranking.NewDefaultRanker(
		featureExtractor,
		scorer,
	)
}
