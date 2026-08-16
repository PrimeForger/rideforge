package builders

import (
	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/candidate/ranking"
	"github.com/ashadashraf/ride-hail-app/internal/config"
)

func BuildRanker(
	cfg *config.Config,
) (ranking.Ranker, error) {

	routingProvider, err := BuildRoutingProvider(cfg)
	if err != nil {
		return nil, err
	}

	travelCalc := ranking.NewDefaultTravelCalculator(routingProvider)

	featureExtractor := ranking.NewDefaultFeatureExtractor(travelCalc)

	scorer := ranking.NewDefaultScorer(&cfg.Ranking)

	return ranking.NewDefaultRanker(
		featureExtractor,
		scorer,
	), nil
}
