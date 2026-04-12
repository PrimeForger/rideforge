package bootstrap

import (
	"github.com/ashadashraf/ride-hail-app/internal/application"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/kafka"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/postgres"
	"github.com/ashadashraf/ride-hail-app/internal/ports"
)

type Container struct {
	RideService        *application.RideService
	MatchingService    *application.MatchingService
	OutboxRepo         ports.OutboxRepository
	ProcessedEventRepo *postgres.ProcessedEventRepository
	TxManager          *postgres.TxManager

	RideProducer  *kafka.Producer
	MatchProducer *kafka.Producer
	DLQProducer   *kafka.Producer
}

func NewContainer() (*Container, error) {

	// --- Database ---
	db, err := postgres.NewDB("postgres://postgres:postgres@localhost:5432/rideforge?sslmode=disable")
	if err != nil {
		return nil, err
	}

	rideRepo := postgres.NewRideRepository(db)
	driverRepo := postgres.NewDriverRepository(db)
	txManager := postgres.NewTxManager(db)
	outboxRepo := postgres.NewOutboxRepository(db)
	processedEventRepo := postgres.NewProcessedEventRepository(db)

	// --- Redis ---
	// eventBus := redisbus.NewEventBus("localhost:6379", "ride-events")

	// -- Kafka Produers ---
	rideProducer := kafka.NewProducer([]string{"localhost:9092"}, "ride.events")
	matchProducer := kafka.NewProducer([]string{"localhost:9092"}, "match.events")
	dlqProducer := kafka.NewProducer([]string{"localhost:9092"}, "ride.events.dlq")

	// ---Application Services ---
	rideService := application.NewRideService(rideRepo, txManager, outboxRepo)
	matchingService := application.NewMatchingService(driverRepo, rideService)
	// idempotencyService := application.NewIdempotencyService(db)

	return &Container{
		RideService:        rideService,
		MatchingService:    matchingService,
		OutboxRepo:         outboxRepo,
		ProcessedEventRepo: processedEventRepo,
		TxManager:          txManager,
		RideProducer:       rideProducer,
		MatchProducer:      matchProducer,
		DLQProducer:        dlqProducer,
	}, nil
}
