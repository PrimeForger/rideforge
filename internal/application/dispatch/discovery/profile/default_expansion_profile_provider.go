package profile

import candidatedensity "github.com/ashadashraf/ride-hail-app/internal/application/dispatch/discovery/density"

type DefaultExpansionProfileProvider struct {
	cfg ExpansionConfig
}

func NewDefaultExpansionProfileProvider(
	cfg ExpansionConfig,
) *DefaultExpansionProfileProvider {
	return &DefaultExpansionProfileProvider{
		cfg: cfg,
	}
}

func (p *DefaultExpansionProfileProvider) Profile(
	density candidatedensity.DensityClass,
) ExpansionProfile {
	switch density {
	case candidatedensity.DensitySparse:
		return p.cfg.Sparse
	case candidatedensity.DensityNormal:
		return p.cfg.Normal
	case candidatedensity.DensityDense:
		return p.cfg.Dense
	default:
		return p.cfg.Default
	}
}
