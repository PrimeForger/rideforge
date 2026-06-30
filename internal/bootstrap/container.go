package bootstrap

import (
	"time"

	"github.com/ashadashraf/ride-hail-app/internal/application"
	"github.com/ashadashraf/ride-hail-app/internal/application/matching"
	"github.com/ashadashraf/ride-hail-app/internal/config"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/geo"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/kafka"
	applogger "github.com/ashadashraf/ride-hail-app/internal/infrastructure/logger"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/postgres"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/realtime"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/redis"
	"github.com/ashadashraf/ride-hail-app/internal/ports"
	"go.uber.org/zap"
)

type Container struct {
	EventRouter *application.EventRouter

	RideService                  *application.RideService
	MatchingEngine               *application.MatchingEngine
	DriverService                *application.DriverService
	DriverResponseService        *application.DriverResponseService
	DriverResponseCommandService *application.DriverResponseCommandService
	DriverLocationService        *application.DriverLocationService
	DriverDeviceService          *application.DriverDeviceService
	DriverMetricsService         *application.DriverMetricsService
	GeoService                   *redis.GeoService
	DriverCache                  *redis.DriverCache

	OutboxRepo         ports.OutboxRepository
	ProcessedEventRepo *postgres.ProcessedEventRepository
	DriverLocker       ports.DriverLocker
	TxManager          *postgres.TxManager

	TimeoutScheduler   *redis.TimeoutScheduler
	RideEventHandler   *matching.RideEventHandler
	DriverOfferHandler *matching.DriverOfferHandler

	RideProducer  *kafka.Producer
	MatchProducer *kafka.Producer
	DLQProducer   *kafka.Producer

	RealtimeHub             *realtime.Hub
	HeartbeatRecoveryWorker *realtime.HeartbeatRecoveryWorker
	Config                  *config.Config

	Logger *zap.Logger
}

func NewContainer() (*Container, error) {
	cfg := config.Load()

	// -- Logger --
	log, err := applogger.New()
	if err != nil {
		return nil, err
	}

	// --- Database ---
	db, err := postgres.NewDB("postgres://postgres:postgres@localhost:5432/rideforge?sslmode=disable")
	if err != nil {
		return nil, err
	}

	// -- Repos --
	rideRepo := postgres.NewRideRepository(db)
	driverRepo := postgres.NewDriverRepository(db)
	outboxRepo := postgres.NewOutboxRepository(db)
	processedEventRepo := postgres.NewProcessedEventRepository(db)
	driverPushTokenRepo := postgres.NewDriverPushTokenRepository(db)
	metricsRepo := postgres.NewDriverMetricsRepository(db)

	// -- Infra --
	txManager := postgres.NewTxManager(db)
	// driverLocker := postgres.NewDBDriverLocker(db)

	// -- Geo --
	h3Service := geo.NewH3Service(cfg.H3.Resolution, cfg.H3.SearchRing)

	// --- Redis ---
	redisClient, err := redis.NewClient("localhost:6379")
	if err != nil {
		return nil, err
	}

	timeoutScheduler := redis.NewTimeoutScheduler(redisClient.GetRaw(), txManager, outboxRepo)
	// eventBus := redisbus.NewEventBus("localhost:6379", "ride-events")
	geoService := redis.NewGeoService(redisClient)
	driverCache := redis.NewDriverCache(redisClient, redis.DriverCacheOptions{
		LocationSeqTTLSeconds: cfg.Realtime.LocationSeqTTLSeconds,
		HeartbeatTTL:          time.Duration(cfg.Realtime.HeartbeatTTLSeconds) * time.Second,
		ConnectionTTL:         time.Duration(cfg.Realtime.ConnectionTTLSeconds) * time.Second,
		DisconnectTTL:         time.Duration(cfg.Realtime.DisconnectTTLSeconds) * time.Second,
		OfferDeliveryTTL:      time.Duration(cfg.Realtime.OfferDeliveryTTLSeconds) * time.Second,
	})
	driverLocker := redis.NewRedisDriverLocker(
		redisClient,
		redis.DriverLockerOptions{
			LockTTL: time.Duration(cfg.Locking.DriverLockTTLSeconds) * time.Second,
		},
	)
	h3Index := redis.NewH3DriverIndex(
		redisClient,
		redis.H3DriverIndexOptions{
			DriverCellTTL: time.Duration(cfg.H3.DriverCellTTLSeconds) * time.Second,
		},
	)

	// -- Kafka Produers ---
	rideProducer := kafka.NewProducer([]string{"localhost:9092"}, "ride.events")
	matchProducer := kafka.NewProducer([]string{"localhost:9092"}, "match.events")
	dlqProducer := kafka.NewProducer([]string{"localhost:9092"}, "ride.events.dlq")

	// -- Real-Time Services ---
	realtimeHub := realtime.NewHub()
	pushNotifier := realtime.NewNoopPushNotifier(driverCache)

	driverOfferGateway := realtime.NewDriverOfferGateway(
		realtimeHub,
		driverCache,
		pushNotifier,
	)

	// --- Application Utilities (Ranking Engine) ---
	rankingEngine := matching.NewRankingEngine(&cfg.Ranking)
	retryPolicy := matching.NewRetryPolicy(cfg.MatchingRetry)

	// --- Application Services ---
	rideService := application.NewRideService(rideRepo, txManager, outboxRepo)
	matchingEngine := application.NewMatchingEngine(driverRepo, driverLocker, outboxRepo, geoService, driverCache, h3Service, h3Index, rankingEngine, cfg, retryPolicy, log)
	driverLocationService := application.NewDriverLocationService(geoService, driverCache, h3Service, h3Index, cfg)
	driverService := application.NewDriverService(driverRepo, txManager, outboxRepo, driverLocationService)
	driverResponseService := application.NewDriverResponseService(rideRepo, driverRepo, driverLocker, outboxRepo, log)
	driverResponseCommandService := application.NewDriverResponseCommandService(txManager, outboxRepo)
	driverDeviceService := application.NewDriverDeviceService(txManager, driverPushTokenRepo, outboxRepo)
	driverMetricsService := application.NewDriverMetricsService(txManager, metricsRepo, driverCache)
	// idempotencyService := application.NewIdempotencyService(db)

	// -- Handlers --
	rideEventHandler := matching.NewRideEventHandler(timeoutScheduler)
	driverOfferHandler := matching.NewDriverOfferHandler(driverCache, timeoutScheduler, driverOfferGateway, cfg, log)

	heartbeatRecoveryWorker := realtime.NewHeartbeatRecoveryWorker(
		driverCache, driverService, time.Duration(cfg.Realtime.HeartbeatRecoveryIntervalSeconds)*time.Second, log,
	)

	eventRouter := application.NewEventRouter(
		txManager,
		processedEventRepo,
		rideService,
		matchingEngine,
		rideEventHandler,
		driverOfferHandler,
		driverResponseService,
		driverMetricsService,
		geoService,
		driverCache,
		h3Index,
		driverLocker,
		log,
	)

	return &Container{
		EventRouter:                  eventRouter,
		RideService:                  rideService,
		MatchingEngine:               matchingEngine,
		DriverService:                driverService,
		DriverResponseService:        driverResponseService,
		DriverResponseCommandService: driverResponseCommandService,
		DriverLocationService:        driverLocationService,
		DriverDeviceService:          driverDeviceService,
		DriverMetricsService:         driverMetricsService,
		GeoService:                   geoService,
		DriverCache:                  driverCache,
		OutboxRepo:                   outboxRepo,
		ProcessedEventRepo:           processedEventRepo,
		DriverLocker:                 driverLocker,
		TxManager:                    txManager,
		TimeoutScheduler:             timeoutScheduler,
		RideEventHandler:             rideEventHandler,
		DriverOfferHandler:           driverOfferHandler,
		RideProducer:                 rideProducer,
		MatchProducer:                matchProducer,
		DLQProducer:                  dlqProducer,
		RealtimeHub:                  realtimeHub,
		HeartbeatRecoveryWorker:      heartbeatRecoveryWorker,
		Config:                       cfg,
		Logger:                       log,
	}, nil
}
