package profile

import candidatedensity "github.com/ashadashraf/ride-hail-app/internal/application/dispatch/discovery/density"

type DefaultExpansionProfileProvider struct {
	maxRing int
}

func NewDefaultExpansionProfileProvider(
	maxRing int,
) *DefaultExpansionProfileProvider {

	return &DefaultExpansionProfileProvider{
		maxRing: maxRing,
	}
}

func (p *DefaultExpansionProfileProvider) Profile(
	density candidatedensity.DensityClass,
) ExpansionProfile {

	switch density {

	case candidatedensity.DensitySparse:

		return ExpansionProfile{
			RingStep:      2,
			MaxRing:       p.maxRing,
			MaxCells:      150,
			MaxCandidates: 150,
		}

	case candidatedensity.DensityNormal:

		return ExpansionProfile{
			RingStep:      1,
			MaxRing:       p.maxRing,
			MaxCells:      80,
			MaxCandidates: 80,
		}

	case candidatedensity.DensityDense:

		return ExpansionProfile{
			RingStep:      1,
			MaxRing:       1,
			MaxCells:      30,
			MaxCandidates: 30,
		}

	default:

		return ExpansionProfile{
			RingStep:      1,
			MaxRing:       p.maxRing,
			MaxCells:      60,
			MaxCandidates: 60,
		}
	}
}
