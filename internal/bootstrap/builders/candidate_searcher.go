package builders

import (
	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/discovery/density"
	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/discovery/expansion"
	profilepipeline "github.com/ashadashraf/ride-hail-app/internal/application/dispatch/discovery/pipeline"
	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/discovery/policy"
	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/discovery/profile"
	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/discovery/refinement"
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
	geoRefiner refinement.Refiner,
) strategy.CandidateSearcher {

	densityClassifier := density.NewDensityClassifier(cfg.Matching.SparseDriverThreshold, cfg.Matching.DenseDriverThreshold)

	expansionProfileProvider := profile.NewDefaultExpansionProfileProvider(profile.ExpansionConfig{
		Sparse: profile.ExpansionProfile{
			RingStep:      cfg.Expansion.SparseRingStep,
			MaxRing:       h3Service.MaxSearchRing(),
			MaxCells:      cfg.Expansion.SparseMaxCells,
			MaxCandidates: cfg.Expansion.SparseMaxCandidates,
		},
		Normal: profile.ExpansionProfile{
			RingStep:      cfg.Expansion.NormalRingStep,
			MaxRing:       h3Service.MaxSearchRing(),
			MaxCells:      cfg.Expansion.NormalMaxCells,
			MaxCandidates: cfg.Expansion.NormalMaxCandidates,
		},
		Dense: profile.ExpansionProfile{
			RingStep:      cfg.Expansion.DenseRingStep,
			MaxRing:       cfg.Expansion.DenseMaxRing,
			MaxCells:      cfg.Expansion.DenseMaxCells,
			MaxCandidates: cfg.Expansion.DenseMaxCandidates,
		},
		Default: profile.ExpansionProfile{
			RingStep:      cfg.Expansion.DefaultRingStep,
			MaxRing:       h3Service.MaxSearchRing(),
			MaxCells:      cfg.Expansion.DefaultMaxCells,
			MaxCandidates: cfg.Expansion.DefaultMaxCandidates,
		},
	})

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

	h3Strategy := strategy.NewH3Strategy(h3Service, ringExpander, budgetFactory, geoRefiner)
	geoStrategy := strategy.NewGeoStrategy(geoService)

	var candidateSearcher strategy.CandidateSearcher

	if cfg.H3.Enabled {
		candidateSearcher = h3Strategy
	} else {
		candidateSearcher = geoStrategy
	}

	return candidateSearcher
}
