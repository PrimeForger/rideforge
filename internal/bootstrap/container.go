package bootstrap

import (
	"time"

	"github.com/ashadashraf/ride-hail-app/internal/application"
	candidatepipeline "github.com/ashadashraf/ride-hail-app/internal/application/dispatch/candidate/pipeline"
	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/candidate/ranking"
	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/candidate/stage"
	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/discovery/density"
	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/discovery/expansion"
	profilepipeline "github.com/ashadashraf/ride-hail-app/internal/application/dispatch/discovery/pipeline"
	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/discovery/policy"
	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/discovery/profile"
	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/discovery/search"
	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/discovery/selector"
	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/discovery/strategy"
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
	densityProvider := redis.NewH3DensityProvider(redisClient, h3Service)

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
	// rankingEngine := matching.NewRankingEngine(&cfg.Ranking)
	retryPolicy := matching.NewRetryPolicy(cfg.MatchingRetry)

	// --- Application Services ---
	rideService := application.NewRideService(rideRepo, txManager, outboxRepo)

	// -----------------------------------------------------------------------------
	// Candidate Discovery
	// -----------------------------------------------------------------------------

	densityClassifier := density.NewDensityClassifier(cfg.Matching.SparseDriverThreshold, cfg.Matching.DenseDriverThreshold)

	expansionProfileProvider := profile.NewDefaultExpansionProfileProvider(h3Service.MaxSearchRing())

	densityRule := policy.NewDensityRule(densityProvider, densityClassifier, expansionProfileProvider)

	profilePipeline := profilepipeline.New(
		densityRule,
	)

	profileSelector := selector.NewDefaultSelector(profilePipeline)

	// expansionPolicy := candidates.NewDefaultRingExpansionPolicy(h3Service.MaxSearchRing())
	ringExpansionPolicy := expansion.NewAdaptiveRingExpansionPolicy(
		profileSelector,
	)

	ringExpander := expansion.NewRingExpander(h3Service, h3Index, ringExpansionPolicy)

	budgetFactory := search.NewDefaultBudgetFactory(h3Service.MaxSearchRing())

	h3Strategy := strategy.NewH3Strategy(h3Service, ringExpander, budgetFactory)
	geoStrategy := strategy.NewGeoStrategy(geoService)

	var candidateSearcher strategy.CandidateSearcher

	if cfg.H3.Enabled {
		candidateSearcher = h3Strategy
	} else {
		candidateSearcher = geoStrategy
	}

	// -----------------------------------------------------------------------------
	// Candidate Ranking
	// -----------------------------------------------------------------------------

	featureExtractor := ranking.NewDefaultFeatureExtractor()

	scorer := ranking.NewDefaultScorer(&cfg.Ranking)

	ranker := ranking.NewDefaultRanker(
		featureExtractor,
		scorer,
	)

	// -----------------------------------------------------------------------------
	// Candidate Pipeline
	// -----------------------------------------------------------------------------

	candidatePipeline := candidatepipeline.New(
		stage.NewDefaultDriverLoader(driverCache),
		stage.NewAvailabilityFilter(),
		stage.NewAlreadyOfferedFilter(driverCache),
		stage.NewRankingStage(ranker),
		stage.NewHeapBuilder(),
	)

	matchingEngine := application.NewMatchingEngine(driverRepo, driverLocker, outboxRepo, candidateSearcher, candidatePipeline, cfg, retryPolicy, log)

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
