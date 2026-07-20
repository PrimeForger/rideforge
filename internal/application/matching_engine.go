package application

import (
	"container/heap"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	appevents "github.com/ashadashraf/ride-hail-app/internal/application/events"
	"github.com/ashadashraf/ride-hail-app/internal/application/matching"
	"github.com/ashadashraf/ride-hail-app/internal/config"
	"github.com/ashadashraf/ride-hail-app/internal/domain/events"
	"github.com/ashadashraf/ride-hail-app/internal/domain/outbox"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/geo"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/observability"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/redis"
	"github.com/ashadashraf/ride-hail-app/internal/ports"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

var matchingTracer = otel.Tracer("application.matching")

type MatchingEngine struct {
	driverRepo  ports.DriverRepository
	locker      ports.DriverLocker
	outboxRepo  ports.OutboxRepository
	geo         *redis.GeoService
	driverCache *redis.DriverCache
	h3          *geo.H3Service
	h3Index     *redis.H3DriverIndex
	ranking     matching.Ranker
	cfg         *config.Config
	retryPolicy *matching.RetryPolicy
	logger      *zap.Logger
}

func NewMatchingEngine(
	driverRepo ports.DriverRepository,
	locker ports.DriverLocker,
	outboxRepo ports.OutboxRepository,
	geo *redis.GeoService,
	driverCache *redis.DriverCache,
	h3 *geo.H3Service,
	h3Index *redis.H3DriverIndex,
	ranking matching.Ranker,
	cfg *config.Config,
	retryPolicy *matching.RetryPolicy,
	logger *zap.Logger,
) *MatchingEngine {
	return &MatchingEngine{
		driverRepo:  driverRepo,
		locker:      locker,
		outboxRepo:  outboxRepo,
		geo:         geo,
		driverCache: driverCache,
		h3:          h3,
		h3Index:     h3Index,
		ranking:     ranking,
		cfg:         cfg,
		retryPolicy: retryPolicy,
		logger:      logger,
	}
}

func (e *MatchingEngine) HandleMatchingStarted(
	ctx context.Context,
	tx *sql.Tx,
	rideID uuid.UUID,
) error {

	ctx, span := matchingTracer.Start(ctx, "MatchingEngine.HandleMatchingStarted")
	defer span.End()

	span.SetAttributes(
		attribute.String("ride.id", rideID.String()),
	)

	start := time.Now()
	result := "success"

	fail := func(metricResult string, err error) error {
		result = metricResult
		span.RecordError(err)
		span.SetAttributes(attribute.String("matching.result", metricResult))
		span.SetStatus(codes.Error, metricResult)
		return err
	}

	defer func() {
		observability.MatchingDurationSeconds.Observe(time.Since(start).Seconds())
		observability.MatchingAttemptsTotal.WithLabelValues(result).Inc()
		span.SetAttributes(attribute.String("matching.final_result", result))
	}()

	// Count attempts
	attemptCount, err := e.driverRepo.CountRideAttemptsTx(ctx, tx, rideID)
	if err != nil {
		e.logger.Error("failed to count matching attempts",
			zap.String("ride_id", rideID.String()),
			zap.Error(err),
		)
		return fail("count_attempts_error", err)
	}

	if attemptCount >= e.cfg.MaxDriverAttempts {
		err := errors.New("max driver attempts reached")

		e.logger.Warn("max matching attempts reached",
			zap.String("ride_id", rideID.String()),
			zap.Int("attempt_count", attemptCount),
			zap.Int("max_attempts", e.cfg.MaxDriverAttempts),
		)

		return fail("max_attempts_reached", err)
	}

	// Dynamic radius expansion
	initialDecision := e.retryPolicy.Decide(matching.RetryInput{
		AttemptCount: attemptCount,
	})

	radius := initialDecision.RadiusKm

	span.SetAttributes(
		attribute.Int("matching.attempt_count", attemptCount),
		attribute.Float64("matching.radius_km", radius),
	)

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
	candidateDriverIDs, err := e.findCandidateDriverIDs(ctx, pickupLat, pickupLng, initialDecision)
	if err != nil {
		e.logger.Error("candidate driver search failed",
			zap.String("ride_id", rideID.String()),
			zap.Error(err),
		)
		return fail("candidate_search_error", err)
	}

	span.SetAttributes(
		attribute.Int("matching.nearby_driver_count", len(candidateDriverIDs)),
	)

	if len(candidateDriverIDs) == 0 {
		err := errors.New("no nearby drivers")

		e.logger.Warn("no nearby drivers found",
			zap.String("ride_id", rideID.String()),
			zap.Float64("radius_km", radius),
		)

		return fail("no_nearby_drivers", err)
	}

	// Batch fetch drivers
	// drivers, err := e.driverRepo.GetEligibleDriversTx(ctx, tx, rideID, nearbyIDs)
	drivers, err := e.driverCache.GetDrivers(ctx, candidateDriverIDs)
	if err != nil {
		e.logger.Error("failed to fetch drivers from cache",
			zap.String("ride_id", rideID.String()),
			zap.Int("nearby_count", len(candidateDriverIDs)),
			zap.Error(err),
		)
		return fail("driver_cache_error", err)
	}

	span.SetAttributes(
		attribute.Int("matching.cached_driver_count", len(drivers)),
	)

	if len(drivers) == 0 {
		err := errors.New("no eligible drivers")

		e.logger.Warn("no drivers found in cache",
			zap.String("ride_id", rideID.String()),
			zap.Int("nearby_count", len(candidateDriverIDs)),
		)

		return fail("no_eligible_drivers", err)
	}

	offeredSet, err := e.driverCache.GetOfferedDrivers(ctx, rideID)
	if err != nil {
		e.logger.Error("failed to fetch offered driver set",
			zap.String("ride_id", rideID.String()),
			zap.Error(err),
		)
		return fail("offered_set_error", err)
	}

	span.SetAttributes(
		attribute.Int("matching.already_offered_count", len(offeredSet)),
	)

	eligibleCount := 0

	for _, d := range drivers {

		if !d.IsAvailable() {
			continue
		}

		if _, exists := offeredSet[d.ID]; exists {
			continue
		}

		eligibleCount++
	}

	decision := e.retryPolicy.Decide(matching.RetryInput{
		AttemptCount:        attemptCount,
		NearbyDriverCount:   len(candidateDriverIDs),
		EligibleDriverCount: eligibleCount,
	})

	e.logger.Info("adaptive retry decision selected",
		zap.String("ride_id", rideID.String()),
		zap.Int("attempt_count", attemptCount),
		zap.Float64("radius_km", decision.RadiusKm),
		zap.Int("offer_batch_size", decision.OfferBatchSize),
		zap.Int("candidate_limit", decision.CandidateLimit),
		zap.Duration("offer_timeout", decision.OfferTimeout),
		zap.Int("nearby_drivers", len(candidateDriverIDs)),
		zap.Int("eligible_drivers", eligibleCount),
	)

	span.SetAttributes(
		attribute.Float64("matching.radius_km", decision.RadiusKm),
		attribute.Int("matching.offer_batch_size", decision.OfferBatchSize),
		attribute.Int("matching.candidate_limit", decision.CandidateLimit),
		attribute.Int64("matching.offer_timeout_ms", decision.OfferTimeout.Milliseconds()),
		attribute.Int("matching.eligible_driver_count", eligibleCount),
	)

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

		distance := matching.HaversineDistanceKm(
			pickupLat,
			pickupLng,
			d.Lat,
			d.Lng,
		)

		score := e.ranking.Score(d, distance)

		heap.Push(h, matching.Candidate{
			DriverID: d.ID,
			Score:    score,
			// Distance: distance,
		})
	}

	span.SetAttributes(
		attribute.Int("matching.candidate_count", h.Len()),
	)

	if h.Len() == 0 {
		err := errors.New("no eligible drivers after filtering")

		e.logger.Warn("no eligible drivers after filtering",
			zap.String("ride_id", rideID.String()),
			zap.Int("cache_driver_count", len(drivers)),
			zap.Int("already_offered_count", len(offeredSet)),
		)

		return fail("no_eligible_after_filtering", err)
	}

	// Offer top N drivers (parallel batch)
	selected := 0

	for h.Len() > 0 && selected < decision.OfferBatchSize {

		candidate := heap.Pop(h).(matching.Candidate)

		ok, err := e.locker.Reserve(ctx, candidate.DriverID, rideID)
		if err != nil {
			e.logger.Error("failed to reserve driver",
				zap.String("ride_id", rideID.String()),
				zap.String("driver_id", candidate.DriverID.String()),
				zap.Error(err),
			)
			return fail("driver_reserve_error", err)
		}

		if !ok {
			observability.DriverOffersTotal.WithLabelValues("reserve_skipped").Inc()
			e.logger.Debug("driver reservation skipped",
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
		if err := e.driverRepo.InsertRideOfferTx(ctx, tx, rideID, candidate.DriverID, attemptCount+1); err != nil {
			e.logger.Error("failed to insert ride driver offer",
				zap.String("ride_id", rideID.String()),
				zap.String("driver_id", candidate.DriverID.String()),
				zap.Error(err),
			)
			e.releaseReservedDriverAfterFailure(ctx, rideID, candidate.DriverID, "insert_offer_failed")

			return fail("insert_offer_error", err)
		}

		event := events.DriverOfferedEvent{
			RideID:          rideID,
			DriverID:        candidate.DriverID,
			OfferTimeoutMs:  decision.OfferTimeout.Milliseconds(),
			MatchingAttempt: attemptCount + 1,
			SearchRadiusKm:  decision.RadiusKm,
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
			e.logger.Error("failed to marshal driver offered event",
				zap.String("ride_id", rideID.String()),
				zap.String("driver_id", candidate.DriverID.String()),
				zap.Error(err),
			)
			e.releaseReservedDriverAfterFailure(ctx, rideID, candidate.DriverID, "marshal_offer_event_failed")

			return fail("marshal_offer_event_error", err)
		}

		if err := e.outboxRepo.Insert(ctx, tx,
			outbox.NewEvent(rideID, envelope.Type, payload),
		); err != nil {
			e.logger.Error("failed to insert driver offered outbox event",
				zap.String("ride_id", rideID.String()),
				zap.String("driver_id", candidate.DriverID.String()),
				zap.Error(err),
			)
			e.releaseReservedDriverAfterFailure(ctx, rideID, candidate.DriverID, "outbox_insert_failed")

			return fail("outbox_insert_error", err)
		}

		observability.DriverOffersTotal.WithLabelValues("offered").Inc()
		selected++
	}

	span.SetAttributes(
		attribute.Int("matching.selected_driver_count", selected),
	)

	if selected == 0 {
		err := errors.New("no drivers reserved")

		e.logger.Warn("matching completed with no drivers reserved",
			zap.String("ride_id", rideID.String()),
		)

		return fail("no_drivers_reserved", err)
	}

	e.logger.Info("matching completed",
		zap.String("ride_id", rideID.String()),
		zap.Int("selected_drivers", selected),
	)

	result = "success"
	span.SetAttributes(attribute.String("matching.result", "success"))

	return nil
}

func (e *MatchingEngine) releaseReservedDriverAfterFailure(
	ctx context.Context,
	rideID uuid.UUID,
	driverID uuid.UUID,
	reason string,
) {
	released, err := e.locker.Release(ctx, driverID, rideID)
	if err != nil {
		e.logger.Error("failed to release reserved driver after matching failure",
			zap.String("ride_id", rideID.String()),
			zap.String("driver_id", driverID.String()),
			zap.String("reason", reason),
			zap.Error(err),
		)
		return
	}

	if !released {
		e.logger.Warn("reserved driver release skipped after matching failure",
			zap.String("ride_id", rideID.String()),
			zap.String("driver_id", driverID.String()),
			zap.String("reason", reason),
		)
		return
	}

	e.logger.Info("reserved driver released after matching failure",
		zap.String("ride_id", rideID.String()),
		zap.String("driver_id", driverID.String()),
		zap.String("reason", reason),
	)
}

func (e *MatchingEngine) findCandidateDriverIDs(
	ctx context.Context,
	pickupLat float64,
	pickupLng float64,
	decision matching.RetryDecision,
) ([]uuid.UUID, error) {
	ctx, span := matchingTracer.Start(ctx, "matching.candidate_search")
	defer span.End()

	if e.cfg.H3.Enabled {

		span.SetAttributes(
			attribute.String("search.backend", "h3"),
			attribute.Int("candidate_limit", decision.CandidateLimit),
		)

		cells, err := e.h3.NeighborCells(pickupLat, pickupLng)
		if err != nil {
			return nil, err
		}

		ids, err := e.h3Index.GetDriversInCells(ctx, cells, decision.CandidateLimit)

		if err != nil {
			return nil, err
		}

		span.SetAttributes(
			attribute.Int("candidate_count", len(ids)),
		)

		return ids, nil
	}

	span.SetAttributes(
		attribute.String("search.backend", "geo"),
		attribute.Float64("radius_km", decision.RadiusKm),
		attribute.Int("candidate_limit", decision.CandidateLimit),
	)

	nearby, err := e.geo.FindNearbyDriversWithDistance(
		ctx,
		pickupLat,
		pickupLng,
		decision.RadiusKm,
		decision.CandidateLimit,
	)

	if err != nil {
		return nil, err
	}

	ids := make([]uuid.UUID, 0, len(nearby))

	for _, d := range nearby {
		ids = append(ids, d.ID)
	}

	span.SetAttributes(
		attribute.Int("candidate_count", len(ids)),
	)

	return ids, nil
}
