package policy

import (
	"context"

	candidatedensity "github.com/ashadashraf/ride-hail-app/internal/application/dispatch/discovery/density"
	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/discovery/profile"
	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/discovery/search"
)

type DensityRule struct {
	densityProvider candidatedensity.DriverDensityProvider
	classifier      *candidatedensity.DensityClassifier

	expansionProvider profile.ExpansionProfileProvider
}

func NewDensityRule(
	densityProvider candidatedensity.DriverDensityProvider,
	classifier *candidatedensity.DensityClassifier,
	expansionProvider profile.ExpansionProfileProvider,
) *DensityRule {

	return &DensityRule{
		densityProvider:   densityProvider,
		classifier:        classifier,
		expansionProvider: expansionProvider,
	}
}

func (r *DensityRule) Apply(
	ctx context.Context,
	state *search.SearchState,
	builder *profile.Builder,
) error {

	driverCount, err := r.densityProvider.DriverCountInRing(
		ctx,
		state.CenterCell,
		state.CurrentRing,
	)

	if err != nil {
		return err
	}

	density := r.classifier.Classify(driverCount)

	state.LastDensity = density

	builder.SetName(
		density.String(),
	)

	expansionProfile := r.expansionProvider.Profile(
		density,
	)

	builder.SetExpansion(
		profile.NewDefaultExpansionPolicy(
			expansionProfile,
		),
	)

	return nil
}
