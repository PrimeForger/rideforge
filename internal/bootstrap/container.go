package bootstrap

import (
	"github.com/ashadashraf/ride-hail-app/internal/application"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/kafka"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/postgres"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/redis"
	"github.com/ashadashraf/ride-hail-app/internal/ports"
)

type Container struct {
	RideService           *application.RideService
	MatchingEngine        *application.MatchingEngine
	DriverResponseService *application.DriverResponseService

	OutboxRepo         ports.OutboxRepository
	ProcessedEventRepo *postgres.ProcessedEventRepository
	TxManager          *postgres.TxManager

	TimeoutScheduler *redis.TimeoutScheduler

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

	// Repos
	rideRepo := postgres.NewRideRepository(db)
	driverRepo := postgres.NewDriverRepository(db)
	outboxRepo := postgres.NewOutboxRepository(db)
	processedEventRepo := postgres.NewProcessedEventRepository(db)

	// Infra
	txManager := postgres.NewTxManager(db)
	driverLocker := postgres.NewDBDriverLocker(db)

	// --- Redis ---
	redisClient, err := redis.NewClient("localhost:6379")
	if err != nil {
		return nil, err
	}

	timeoutScheduler := redis.NewTimeoutScheduler(redisClient.GetRaw(), txManager, outboxRepo)
	// eventBus := redisbus.NewEventBus("localhost:6379", "ride-events")

	// -- Kafka Produers ---
	rideProducer := kafka.NewProducer([]string{"localhost:9092"}, "ride.events")
	matchProducer := kafka.NewProducer([]string{"localhost:9092"}, "match.events")
	dlqProducer := kafka.NewProducer([]string{"localhost:9092"}, "ride.events.dlq")

	// ---Application Services ---
	rideService := application.NewRideService(rideRepo, txManager, outboxRepo)
	matchingEngine := application.NewMatchingEngine(driverRepo, driverLocker, outboxRepo)
	driverResponseService := application.NewDriverResponseService(rideRepo, driverRepo, driverLocker, outboxRepo, timeoutScheduler)
	// idempotencyService := application.NewIdempotencyService(db)

	return &Container{
		RideService:           rideService,
		MatchingEngine:        matchingEngine,
		DriverResponseService: driverResponseService,
		OutboxRepo:            outboxRepo,
		ProcessedEventRepo:    processedEventRepo,
		TxManager:             txManager,
		TimeoutScheduler:      timeoutScheduler,
		RideProducer:          rideProducer,
		MatchProducer:         matchProducer,
		DLQProducer:           dlqProducer,
	}, nil
}
