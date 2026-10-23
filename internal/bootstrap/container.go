package bootstrap

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ashadashraf/ride-hail-app/internal/application"
	applicationdemand "github.com/ashadashraf/ride-hail-app/internal/application/demand"
	applicationdriver "github.com/ashadashraf/ride-hail-app/internal/application/driver"
	applicationlocality "github.com/ashadashraf/ride-hail-app/internal/application/locality"
	applicationlocation "github.com/ashadashraf/ride-hail-app/internal/application/location"
	"github.com/ashadashraf/ride-hail-app/internal/application/matching"
	applicationoccupancy "github.com/ashadashraf/ride-hail-app/internal/application/occupancy"
	applicationsurge "github.com/ashadashraf/ride-hail-app/internal/application/surge"
	"github.com/ashadashraf/ride-hail-app/internal/bootstrap/builders"
	"github.com/ashadashraf/ride-hail-app/internal/config"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/geo"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/kafka"
	applogger "github.com/ashadashraf/ride-hail-app/internal/infrastructure/logger"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/observability"
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
	DriverService                *applicationdriver.DriverService
	DriverResponseService        *applicationdriver.DriverResponseService
	DriverResponseCommandService *applicationdriver.DriverResponseCommandService
	DriverLocationService        *applicationlocation.DriverLocationService
	DriverDeviceService          *applicationdriver.DriverDeviceService
	DriverMetricsService         *applicationdriver.DriverMetricsService
	H3RecoveryService            *applicationlocation.H3RecoveryService
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

	RealtimeHub                *realtime.Hub
	HeartbeatRecoveryWorker    *realtime.HeartbeatRecoveryWorker
	H3ReconciliationWorker     *applicationlocation.H3ReconciliationWorker
	DemandReconciliationWorker *applicationdemand.DemandReconciliationWorker
	DemandStore                ports.CellDemandStore
	DemandService              *applicationdemand.DemandService
	DemandRecoveryService      *applicationdemand.DemandRecoveryService
	LocalityService            *applicationlocality.LocalityService
	LocalityRepository         *postgres.LocalityRepository
	SurgePolicyRepository      *postgres.SurgePolicyRepository
	SurgePolicyCache           *applicationsurge.SurgePolicyCache
	SurgePolicyRefreshWorker   *applicationsurge.SurgePolicyRefreshWorker
	SurgeService               *applicationsurge.SurgeService
	OccupancyStore             ports.CellOccupancyStore
	OccupancyService           *applicationoccupancy.OccupancyService
	Config                     *config.Config

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
	if strings.TrimSpace(cfg.DatabaseURL) == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	dbCtx, cancelDB := context.WithTimeout(context.Background(), cfg.DatabaseConnectTimeout)
	defer cancelDB()
	db, err := postgres.NewDB(dbCtx, cfg.DatabaseURL)
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

	// -- Geo --
	h3Service := geo.NewH3Service(cfg.H3.Resolution, cfg.H3.SearchRing)
	localityRepo := postgres.NewLocalityRepository(db, h3Service)
	surgePolicyRepo := postgres.NewSurgePolicyRepository(db)

	// --- Redis ---
	redisClient, err := redis.NewClient("localhost:6379")
	if err != nil {
		return nil, err
	}

	timeoutScheduler := redis.NewTimeoutScheduler(redisClient.GetRaw(), txManager, outboxRepo)
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
	demandStore := redis.NewRedisDemandStore(
		redisClient,
		redis.DemandStoreOptions{
			DemandTTL: time.Duration(cfg.Demand.DemandTTLSeconds) * time.Second,
		},
	)
	occupancyStore := redis.NewRedisOccupancyStore(
		redisClient,
		redis.OccupancyStoreOptions{
			WarmImbalanceThreshold: cfg.Occupancy.WarmImbalanceThreshold,
			HotImbalanceThreshold:  cfg.Occupancy.HotImbalanceThreshold,
			MinDemandForHot:        cfg.Occupancy.MinDemandForHot,
			DemandTTL:              time.Duration(cfg.Demand.DemandTTLSeconds) * time.Second,
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
	retryPolicy := matching.NewRetryPolicy(cfg.MatchingRetry)

	// --- Application Services ---
	rideService := application.NewRideService(rideRepo, txManager, outboxRepo)

	// -----------------------------------------------------------------------------
	// Candidate Discovery
	// -----------------------------------------------------------------------------

	geoRefiner := builders.BuildGeoRefiner(geoService)

	candidateSearcher := builders.BuildCandidateSearcher(
		cfg,
		h3Service,
		h3Index,
		geoService,
		densityProvider,
		geoRefiner,
	)

	// -----------------------------------------------------------------------------
	// Candidate Ranking
	// -----------------------------------------------------------------------------

	ranker, err := builders.BuildRanker(cfg)
	if err != nil {
		return nil, err
	}

	// -----------------------------------------------------------------------------
	// Candidate Pipeline
	// -----------------------------------------------------------------------------

	candidatePipeline := builders.BuildCandidatePipeline(
		driverCache,
		driverCache,
		ranker,
	)

	matchingEngine := application.NewMatchingEngine(driverRepo, driverLocker, outboxRepo, candidateSearcher, candidatePipeline, cfg, retryPolicy, log)

	driverLocationService := applicationlocation.NewDriverLocationService(geoService, driverCache, h3Service, h3Index, cfg)
	driverService := applicationdriver.NewDriverService(driverRepo, txManager, outboxRepo, driverLocationService)
	driverResponseService := applicationdriver.NewDriverResponseService(rideRepo, driverRepo, driverLocker, outboxRepo, log)
	driverResponseCommandService := applicationdriver.NewDriverResponseCommandService(txManager, outboxRepo)
	driverDeviceService := applicationdriver.NewDriverDeviceService(txManager, driverPushTokenRepo, outboxRepo)
	driverMetricsService := applicationdriver.NewDriverMetricsService(txManager, metricsRepo, driverCache)
	h3RecoveryService := applicationlocation.NewH3RecoveryService(driverRepo, h3Service, h3Index, geoService, driverCache, log)

	demandService := applicationdemand.NewDemandService(demandStore, h3Service, log)
	demandRecoveryService := applicationdemand.NewDemandRecoveryService(rideRepo, h3Service, demandStore, log)

	occupancyService := applicationoccupancy.NewOccupancyService(occupancyStore, h3Service, log)
	localityService := applicationlocality.NewLocalityService(localityRepo, localityRepo)
	surgeMetrics := observability.SurgeMetrics{}
	surgePolicyCache := applicationsurge.NewSurgePolicyCache(
		surgePolicyRepo,
		time.Duration(cfg.SurgePolicy.MaxStalenessSeconds)*time.Second,
	)
	surgePolicyRefreshWorker := applicationsurge.NewSurgePolicyRefreshWorker(
		surgePolicyCache,
		time.Duration(cfg.SurgePolicy.RefreshIntervalSeconds)*time.Second,
		surgeMetrics,
		log,
	)
	surgeService := applicationsurge.NewSurgeService(h3Service, localityRepo, surgePolicyCache, occupancyService, surgeMetrics, log)

	// -- Handlers --
	rideEventHandler := matching.NewRideEventHandler(timeoutScheduler)
	driverOfferHandler := matching.NewDriverOfferHandler(driverCache, timeoutScheduler, driverOfferGateway, cfg, log)

	heartbeatRecoveryWorker := realtime.NewHeartbeatRecoveryWorker(
		driverCache, driverService, time.Duration(cfg.Realtime.HeartbeatRecoveryIntervalSeconds)*time.Second, log,
	)
	h3ReconciliationWorker := applicationlocation.NewH3ReconciliationWorker(
		h3RecoveryService,
		cfg.H3.ReconciliationEnabled,
		time.Duration(cfg.H3.ReconciliationIntervalSeconds)*time.Second,
		log,
	)
	demandReconciliationWorker := applicationdemand.NewDemandReconciliationWorker(
		demandRecoveryService,
		occupancyService,
		cfg.Demand.ReconciliationEnabled,
		time.Duration(cfg.Demand.ReconciliationIntervalSeconds)*time.Second,
		log,
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
		demandService,
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
		H3RecoveryService:            h3RecoveryService,
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
		H3ReconciliationWorker:       h3ReconciliationWorker,
		DemandReconciliationWorker:   demandReconciliationWorker,
		DemandStore:                  demandStore,
		DemandService:                demandService,
		DemandRecoveryService:        demandRecoveryService,
		OccupancyStore:               occupancyStore,
		OccupancyService:             occupancyService,
		LocalityService:              localityService,
		LocalityRepository:           localityRepo,
		SurgePolicyRepository:        surgePolicyRepo,
		SurgePolicyCache:             surgePolicyCache,
		SurgePolicyRefreshWorker:     surgePolicyRefreshWorker,
		SurgeService:                 surgeService,
		Config:                       cfg,
		Logger:                       log,
	}, nil
}
