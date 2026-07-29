package profile

import "github.com/ashadashraf/ride-hail-app/internal/application/dispatch/discovery/density"

type ExpansionProfileProvider interface {
	Profile(density density.DensityClass) ExpansionProfile
}
