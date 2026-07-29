package expansion

import (
	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/discovery/density"
)

type RingDecision struct {
	// Ring to execute in this iteration.
	Ring int

	// Ring to execute on the next iteration.
	NextRing int

	// Observability / future dispatch decisions.
	Density density.DensityClass
}
