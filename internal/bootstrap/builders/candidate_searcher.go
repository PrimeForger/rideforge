package builders

import (
	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/discovery/density"
	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/discovery/expansion"
	profilepipeline "github.com/ashadashraf/ride-hail-app/internal/application/dispatch/discovery/pipeline"
	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/discovery/policy"
	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/discovery/profile"
	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/discovery/search"
	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/discovery/selector"
	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/discovery/strategy"
	"github.com/ashadashraf/ride-hail-app/internal/config"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/geo"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/redis"
)

func BuildCandidateSearcher(
	cfg *config.Config,
	h3Service *geo.H3Service,
	h3Index *redis.H3DriverIndex,
	geoService *redis.GeoService,
	densityProvider density.DriverDensityProvider,
) strategy.CandidateSearcher {

	densityClassifier := density.NewDensityClassifier(cfg.Matching.SparseDriverThreshold, cfg.Matching.DenseDriverThreshold)

	expansionProfileProvider := profile.NewDefaultExpansionProfileProvider(h3Service.MaxSearchRing())

	densityRule := policy.NewDensityRule(densityProvider, densityClassifier, expansionProfileProvider)

	profilePipeline := profilepipeline.New(
		densityRule,
	)

	profileSelector := selector.NewDefaultSelector(profilePipeline)

	// expansionPolicy := candidates.NewDefaultRingExpansionPolicy(h3Service.MaxSearchRing())
	ringExpansionPolicy := expansion.NewAdaptiveRingExpansionPolicy(
		profileSelector,
	)

	ringExpander := expansion.NewRingExpander(h3Service, h3Index, ringExpansionPolicy)

	budgetFactory := search.NewDefaultBudgetFactory(h3Service.MaxSearchRing())

	h3Strategy := strategy.NewH3Strategy(h3Service, ringExpander, budgetFactory)
	geoStrategy := strategy.NewGeoStrategy(geoService)

	var candidateSearcher strategy.CandidateSearcher

	if cfg.H3.Enabled {
		candidateSearcher = h3Strategy
	} else {
		candidateSearcher = geoStrategy
	}

	return candidateSearcher
}
