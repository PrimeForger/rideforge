package profile

import "github.com/ashadashraf/ride-hail-app/internal/application/dispatch/discovery/search"

type ExpansionPolicy interface {
	ConfigureBudget(
		budget *search.SearchBudget,
	)

	NextRing(
		currentRing int,
	) int
}
