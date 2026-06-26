package application

import (
	"container/heap"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"math"
	"time"

	appevents "github.com/ashadashraf/ride-hail-app/internal/application/events"
	"github.com/ashadashraf/ride-hail-app/internal/application/matching"
	"github.com/ashadashraf/ride-hail-app/internal/config"
	"github.com/ashadashraf/ride-hail-app/internal/domain/events"
	"github.com/ashadashraf/ride-hail-app/internal/domain/outbox"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/observability"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/redis"
	"github.com/ashadashraf/ride-hail-app/internal/ports"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type MatchingEngine struct {
	driverRepo  ports.DriverRepository
	locker      ports.DriverLocker
	outboxRepo  ports.OutboxRepository
	geo         *redis.GeoService
	driverCache *redis.DriverCache
	ranking     matching.Ranker
	cfg         *config.Config
	logger      *zap.Logger
}

func NewMatchingEngine(
	driverRepo ports.DriverRepository,
	locker ports.DriverLocker,
	outboxRepo ports.OutboxRepository,
	geo *redis.GeoService,
	driverCache *redis.DriverCache,
	ranking matching.Ranker,
	cfg *config.Config,
	logger *zap.Logger,
) *MatchingEngine {
	return &MatchingEngine{
		driverRepo:  driverRepo,
		locker:      locker,
		outboxRepo:  outboxRepo,
		geo:         geo,
		driverCache: driverCache,
		ranking:     ranking,
		cfg:         cfg,
		logger:      logger,
	}
}

func (e *MatchingEngine) HandleMatchingStarted(
	ctx context.Context,
	tx *sql.Tx,
	rideID uuid.UUID,
) error {
	start := time.Now()
	result := "success"

	defer func() {
		observability.MatchingDurationSeconds.Observe(time.Since(start).Seconds())
		observability.MatchingAttemptsTotal.WithLabelValues(result).Inc()
	}()

	// Count attempts
	attemptCount, err := e.driverRepo.CountRideAttemptsTx(ctx, tx, rideID)
	if err != nil {
		result = "count_attempts_error"
		e.logger.Error("failed to count matching attempts",
			zap.String("ride_id", rideID.String()),
			zap.Error(err),
		)
		return err
	}

	if attemptCount >= e.cfg.MaxDriverAttempts {
		result = "max_attempts_reached"
		e.logger.Warn("max matching attempts reached",
			zap.String("ride_id", rideID.String()),
			zap.Int("attempt_count", attemptCount),
			zap.Int("max_attempts", e.cfg.MaxDriverAttempts),
		)
		return errors.New("max driver attempts reached")
	}

	// Dynamic radius expansion
	radius := e.computeRadius(attemptCount)

	e.logger.Info("matching started",
		zap.String("ride_id", rideID.String()),
		zap.Int("attempt_count", attemptCount),
		zap.Float64("radius_km", radius),
	)

	// 2. Get available drivers excluding already tried
	// drivers, err := e.driverRepo.GetAvailableDriversExcludingTx(ctx, tx, rideID)
	// if err != nil {
	// 	return err
	// }

	// TODO: replace with real pickup location
	pickupLat := 17.3850
	pickupLng := 78.4867

	// Get nearby driver IDs
	nearby, err := e.geo.FindNearbyDriversWithDistance(ctx, pickupLat, pickupLng, radius, 50)
	if err != nil {
		result = "geo_search_error"
		e.logger.Error("geo driver search failed",
			zap.String("ride_id", rideID.String()),
			zap.Float64("radius_km", radius),
			zap.Error(err),
		)
		return err
	}

	if len(nearby) == 0 {
		result = "no_nearby_drivers"
		e.logger.Warn("no nearby drivers found",
			zap.String("ride_id", rideID.String()),
			zap.Float64("radius_km", radius),
		)
		return errors.New("no nearby drivers")
	}

	ids := make([]uuid.UUID, 0, len(nearby))
	distanceMap := make(map[uuid.UUID]float64)

	for _, d := range nearby {
		ids = append(ids, d.ID)
		distanceMap[d.ID] = d.Distance
	}

	// Batch fetch drivers
	// drivers, err := e.driverRepo.GetEligibleDriversTx(ctx, tx, rideID, nearbyIDs)
	drivers, err := e.driverCache.GetDrivers(ctx, ids)
	if err != nil {
		result = "driver_cache_error"
		e.logger.Error("failed to fetch drivers from cache",
			zap.String("ride_id", rideID.String()),
			zap.Int("nearby_count", len(nearby)),
			zap.Error(err),
		)
		return err
	}

	if len(drivers) == 0 {
		result = "no_eligible_drivers"
		e.logger.Warn("no drivers found in cache",
			zap.String("ride_id", rideID.String()),
			zap.Int("nearby_count", len(nearby)),
		)
		return errors.New("no eligible drivers")
	}

	offeredSet, err := e.driverCache.GetOfferedDrivers(ctx, rideID)
	if err != nil {
		result = "offered_set_error"
		e.logger.Error("failed to fetch offered driver set",
			zap.String("ride_id", rideID.String()),
			zap.Error(err),
		)
		return err
	}

	// Build heap
	h := &matching.MaxHeap{}
	heap.Init(h)

	for _, d := range drivers {

		if !d.IsAvailable() {
			continue
		}

		if _, exists := offeredSet[d.ID]; exists {
			continue
		}

		// distance := e.geo.Distance(ctx, pickupLat, pickupLng, driver.ID)
		// distance := haversineDistance(pickupLat, pickupLng, driver.Lat, driver.Lng)

		distance, ok := distanceMap[d.ID]
		if !ok {
			continue
		}

		score := e.ranking.Score(d, distance)

		heap.Push(h, matching.Candidate{
			DriverID: d.ID,
			Score:    score,
			// Distance: distance,
		})
	}

	if h.Len() == 0 {
		result = "no_eligible_after_filtering"
		e.logger.Warn("no eligible drivers after filtering",
			zap.String("ride_id", rideID.String()),
			zap.Int("cache_driver_count", len(drivers)),
			zap.Int("already_offered_count", len(offeredSet)),
		)
		return errors.New("no eligible drivers after filtering")
	}

	// Offer top N drivers (parallel batch)
	selected := 0

	for h.Len() > 0 && selected < e.cfg.OfferBatchSize {

		candidate := heap.Pop(h).(matching.Candidate)

		ok, err := e.locker.ReserveTx(ctx, tx, candidate.DriverID, rideID)
		if err != nil {
			result = "driver_reserve_error"
			e.logger.Error("failed to reserve driver",
				zap.String("ride_id", rideID.String()),
				zap.String("driver_id", candidate.DriverID.String()),
				zap.Error(err),
			)
			return err
		}

		if !ok {
			observability.DriverOffersTotal.WithLabelValues("reserve_skipped").Inc()
			e.logger.Warn("driver reservation skipped",
				zap.String("ride_id", rideID.String()),
				zap.String("driver_id", candidate.DriverID.String()),
				zap.Float64("score", candidate.Score),
			)
			continue
		}

		e.logger.Info("driver reserved for offer",
			zap.String("ride_id", rideID.String()),
			zap.String("driver_id", candidate.DriverID.String()),
			zap.Int("attempt", attemptCount+1),
			zap.Float64("score", candidate.Score),
		)

		// Record attempt in DB
		err = e.driverRepo.InsertRideOfferTx(ctx, tx, rideID, candidate.DriverID, attemptCount+1)
		if err != nil {
			result = "insert_offer_error"
			e.logger.Error("failed to insert ride driver offer",
				zap.String("ride_id", rideID.String()),
				zap.String("driver_id", candidate.DriverID.String()),
				zap.Error(err),
			)
			return err
		}

		event := events.DriverOfferedEvent{
			RideID:   rideID,
			DriverID: candidate.DriverID,
		}

		envelope := appevents.Envelope{
			ID:        uuid.NewString(),
			Type:      event.Name(),
			Aggregate: rideID.String(),
			Data:      event,
			Occurred:  time.Now(),
		}

		payload, err := json.Marshal(envelope)
		if err != nil {
			result = "marshal_offer_event_error"
			e.logger.Error("failed to marshal driver offered event",
				zap.String("ride_id", rideID.String()),
				zap.String("driver_id", candidate.DriverID.String()),
				zap.Error(err),
			)
			return err
		}

		if err := e.outboxRepo.Insert(ctx, tx,
			outbox.NewEvent(rideID, envelope.Type, payload),
		); err != nil {
			result = "outbox_insert_error"
			e.logger.Error("failed to insert driver offered outbox event",
				zap.String("ride_id", rideID.String()),
				zap.String("driver_id", candidate.DriverID.String()),
				zap.Error(err),
			)
			return err
		}

		observability.DriverOffersTotal.WithLabelValues("offered").Inc()
		selected++
	}

	if selected == 0 {
		result = "no_drivers_reserved"
		e.logger.Warn("matching completed with no drivers reserved",
			zap.String("ride_id", rideID.String()),
			zap.Int("candidate_count", h.Len()),
		)
		return errors.New("no drivers reserved")
	}

	e.logger.Info("matching completed",
		zap.String("ride_id", rideID.String()),
		zap.Int("selected_drivers", selected),
	)

	result = "success"
	return nil
}

func (e *MatchingEngine) computeRadius(attempt int) float64 {
	return e.cfg.SearchRadiusKm * math.Pow(2, float64(attempt))
}
