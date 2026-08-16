package builders

import (
	"fmt"

	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/routing"
	"github.com/ashadashraf/ride-hail-app/internal/config"
)

func BuildRoutingProvider(
	cfg *config.Config,
) (routing.Provider, error) {

	switch cfg.Routing.Provider {
	case "", "none":
		return routing.NewDisabledProvider(), nil

	default:
		return nil, fmt.Errorf("unsupported routing provider %q: only \"none\" is currently supported", cfg.Routing.Provider)
	}
}
