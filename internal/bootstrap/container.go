package bootstrap

import (
	"github.com/ashadashraf/ride-hail-app/internal/application"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/postgres"
	redisbus "github.com/ashadashraf/ride-hail-app/internal/infrastructure/redis"
)

type Container struct {
	RideService     *application.RideService
	MatchingService *application.MatchingService
}

func NewContainer() (*Container, error) {

	// --- Database ---
	db, err := postgres.NewDB("postgres://postgres:postgres@localhost:5432/rideforge?sslmode=disable")
	if err != nil {
		return nil, err
	}

	rideRepo := postgres.NewRideRepository(db)
	driverRepo := postgres.NewDriverRepository(db)

	// --- Redis ---
	eventBus := redisbus.NewEventBus("localhost:6379", "ride-events")

	// ---Application Services ---
	rideService := application.NewRideService(rideRepo, eventBus)
	matchingService := application.NewMatchingService(driverRepo, rideService)

	return &Container{
		RideService:     rideService,
		MatchingService: matchingService,
	}, nil
}
