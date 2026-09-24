package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/candidate"
	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/candidate/pipeline"
	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/discovery/search"
	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/discovery/strategy"
	appevents "github.com/ashadashraf/ride-hail-app/internal/application/events"
	"github.com/ashadashraf/ride-hail-app/internal/application/matching"
	"github.com/ashadashraf/ride-hail-app/internal/config"
	"github.com/ashadashraf/ride-hail-app/internal/domain/events"
	"github.com/ashadashraf/ride-hail-app/internal/domain/outbox"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/observability"
	"github.com/ashadashraf/ride-hail-app/internal/ports"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

var matchingTracer = otel.Tracer("application.matching")

type MatchingEngine struct {
	driverRepo ports.DriverRepository
	locker     ports.DriverLocker
	outboxRepo ports.OutboxRepository

	candidateSearcher strategy.CandidateSearcher
	candidatePipeline pipeline.Pipeline

	cfg         *config.Config
	retryPolicy *matching.RetryPolicy
	logger      *zap.Logger
}

func NewMatchingEngine(
	driverRepo ports.DriverRepository,
	locker ports.DriverLocker,
	outboxRepo ports.OutboxRepository,
	candidateSearcher strategy.CandidateSearcher,
	candidatePipeline pipeline.Pipeline,
	cfg *config.Config,
	retryPolicy *matching.RetryPolicy,
	logger *zap.Logger,
) *MatchingEngine {
	return &MatchingEngine{
		driverRepo:        driverRepo,
		locker:            locker,
		outboxRepo:        outboxRepo,
		candidateSearcher: candidateSearcher,
		candidatePipeline: candidatePipeline,
		cfg:               cfg,
		retryPolicy:       retryPolicy,
		logger:            logger,
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

	// -----------------------------------------------------------------------------
	// Phase 1: Determine matching attempt and derive the initial dispatch strategy.
	// -----------------------------------------------------------------------------

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

	span.SetAttributes(
		attribute.Int("matching.attempt_count", attemptCount),
		attribute.Float64("matching.initial_radius_km", initialDecision.RadiusKm),
	)

	e.logger.Info("matching started",
		zap.String("ride_id", rideID.String()),
		zap.Int("attempt_count", attemptCount),
		zap.Float64("radius_km", initialDecision.RadiusKm),
	)

	// TODO: replace with real pickup location
	pickupLat := 17.3850
	pickupLng := 78.4867

	// -----------------------------------------------------------------------------
	// Phase 2: Discover nearby candidate drivers using the configured search
	// strategy (H3, Geo, etc.). This stage performs discovery only and does not
	// load driver details or rank candidates.
	// -----------------------------------------------------------------------------

	lookupStart := time.Now()
	lookupCtx, lookupSpan := matchingTracer.Start(ctx, "dispatch.candidate_lookup")

	discoveryResult, err := e.candidateSearcher.FindCandidates(
		lookupCtx,
		search.Request{
			PickupLat: pickupLat,
			PickupLng: pickupLng,

			RadiusKm: initialDecision.RadiusKm,

			CandidateLimit: initialDecision.CandidateLimit,

			MatchingAttempt: attemptCount,
		},
	)
	observability.DispatchCandidateLookupDurationSeconds.Observe(time.Since(lookupStart).Seconds())

	if err != nil {
		observability.CandidateRetrievalOutcomeTotal.WithLabelValues("error").Inc()
		lookupSpan.SetAttributes(attribute.String("search.result", "error"))
		lookupSpan.End()

		e.logger.Error("candidate driver search failed",
			zap.String("ride_id", rideID.String()),
			zap.Error(err),
		)
		return fail("candidate_search_error", err)
	}

	discoveredCount := 0
	if discoveryResult.Candidates != nil {
		discoveredCount = discoveryResult.Candidates.Len()
	}

	observability.H3LookupCellsTotal.Observe(float64(discoveryResult.CellsVisited))
	observability.H3SearchRingsVisited.Observe(float64(discoveryResult.RingsVisited))
	observability.H3LookupResultsTotal.Observe(float64(discoveredCount))

	// -----------------------------------------------------------------------------
	// Phase 3: Execute the candidate pipeline.
	//
	// Responsibilities:
	//   - Load driver data
	//   - Filter unavailable drivers
	//   - Filter already offered drivers
	//   - Rank candidates
	//   - Build ordered candidate source
	//
	// After this point MatchingEngine consumes candidates without knowing how
	// they were filtered or ranked.
	// -----------------------------------------------------------------------------

	pipelineCtx := pipeline.NewContext(
		rideID,
		pickupLat,
		pickupLng,
		attemptCount,
		initialDecision.RadiusKm,
		initialDecision.CandidateLimit,
	)

	rankingStart := time.Now()
	rankingCtx, rankingSpan := matchingTracer.Start(ctx, "dispatch.ranking")

	err = e.candidatePipeline.Execute(
		rankingCtx,
		pipelineCtx,
		discoveryResult.Candidates,
	)
	rankingSpan.End()
	observability.DispatchRankingDurationSeconds.Observe(time.Since(rankingStart).Seconds())

	if err != nil {
		observability.CandidateRetrievalOutcomeTotal.WithLabelValues("error").Inc()
		lookupSpan.SetAttributes(attribute.String("search.result", "error"))
		lookupSpan.End()

		e.logger.Error("candidate pipeline execution failed",
			zap.String("ride_id", rideID.String()),
			zap.Error(err),
		)
		return fail("candidate_pipeline_error", err)
	}

	observability.CandidatePipelineCount.WithLabelValues("loaded").Observe(float64(pipelineCtx.Result.LoadedCandidates))
	observability.CandidatePipelineCount.WithLabelValues("filtered").Observe(float64(pipelineCtx.Result.FilteredCandidates))
	observability.CandidatePipelineCount.WithLabelValues("ranked").Observe(float64(pipelineCtx.Result.RankedCandidates))

	retrievalOutcome := "found"
	if pipelineCtx.Result.RankedCandidates == 0 {
		retrievalOutcome = "empty"
	}
	observability.CandidateRetrievalOutcomeTotal.WithLabelValues(retrievalOutcome).Inc()

	lookupSpan.SetAttributes(
		attribute.String("search.backend", discoveryResult.Backend),
		attribute.Int("search.rings_visited", discoveryResult.RingsVisited),
		attribute.Int("search.cells_visited", discoveryResult.CellsVisited),
		attribute.Int("candidates.discovered", discoveredCount),
		attribute.Int("candidates.loaded", pipelineCtx.Result.LoadedCandidates),
		attribute.Int("candidates.filtered", pipelineCtx.Result.FilteredCandidates),
		attribute.Int("candidates.ranked", pipelineCtx.Result.RankedCandidates),
		attribute.String("search.result", retrievalOutcome),
	)
	lookupSpan.End()

	span.SetAttributes(
		attribute.String("matching.discovery_backend", discoveryResult.Backend),
		attribute.Int("matching.cells_visited", discoveryResult.CellsVisited),
		attribute.Int("matching.rings_visited", discoveryResult.RingsVisited),
		attribute.Float64("matching.discovery_radius_km", discoveryResult.RadiusKm),
		attribute.Int("matching.loaded_candidates", pipelineCtx.Result.LoadedCandidates),
		attribute.Int("matching.filtered_candidates", pipelineCtx.Result.FilteredCandidates),
		attribute.Int("matching.ranked_candidates", pipelineCtx.Result.RankedCandidates),
	)

	// -----------------------------------------------------------------------------
	// Phase 4: Recompute the dispatch strategy using pipeline statistics.
	// The initial strategy is based only on retry count.
	// This strategy also considers actual candidate availability.
	// -----------------------------------------------------------------------------

	dispatchDecision := e.retryPolicy.Decide(matching.RetryInput{
		AttemptCount: attemptCount,

		NearbyDriverCount:   pipelineCtx.Result.LoadedCandidates,
		EligibleDriverCount: pipelineCtx.Result.RankedCandidates,

		PreviouslyOfferedCount:  attemptCount,
		PreviouslyAcceptedCount: 0,
		PreviouslyRejectedCount: 0,
		PreviouslyTimedOutCount: 0,
	})

	e.logger.Info("adaptive retry decision selected",
		zap.String("ride_id", rideID.String()),
		zap.Int("attempt_count", attemptCount),
		zap.Float64("radius_km", dispatchDecision.RadiusKm),
		zap.Int("offer_batch_size", dispatchDecision.OfferBatchSize),
		zap.Int("candidate_limit", dispatchDecision.CandidateLimit),
		zap.Duration("offer_timeout", dispatchDecision.OfferTimeout),
		zap.Int("loaded_candidates", pipelineCtx.Result.LoadedCandidates),
		zap.Int("filtered_candidates", pipelineCtx.Result.FilteredCandidates),
		zap.Int("ranked_candidates", pipelineCtx.Result.RankedCandidates),
	)

	span.SetAttributes(
		attribute.Float64("matching.offer_radius_km", dispatchDecision.RadiusKm),
		attribute.Int("matching.offer_batch_size", dispatchDecision.OfferBatchSize),
		attribute.Int("matching.candidate_limit", dispatchDecision.CandidateLimit),
		attribute.Int64("matching.offer_timeout_ms", dispatchDecision.OfferTimeout.Milliseconds()),
	)

	// -----------------------------------------------------------------------------
	// Phase 5: Offer drivers in ranking order until the configured batch size
	// has been reached.
	//
	// Candidate ordering is already determined by the pipeline.
	// MatchingEngine only reserves, persists and publishes offers.
	// -----------------------------------------------------------------------------

	offered := 0

	for {
		candidate, ok := pipelineCtx.Result.Candidates.Next()
		if !ok || offered >= dispatchDecision.OfferBatchSize {
			break
		}

		offeredNow, err := e.offerCandidate(
			ctx,
			tx,
			rideID,
			attemptCount+1,
			dispatchDecision,
			candidate,
		)
		if err != nil {
			return fail("offer_candidate_error", err)
		}

		if offeredNow {
			offered++
		}
	}

	span.SetAttributes(
		attribute.Int("matching.offered_driver_count", offered),
		attribute.Bool("matching.success", offered > 0),
	)

	// -----------------------------------------------------------------------------
	// Phase 6: Complete matching.
	//
	// Matching succeeds if at least one driver has been offered.
	// Retry orchestration happens outside this handler.
	// -----------------------------------------------------------------------------

	if offered == 0 {
		err := errors.New("no drivers reserved")

		e.logger.Warn("matching completed with no drivers reserved",
			zap.String("ride_id", rideID.String()),
		)

		return fail("no_drivers_reserved", err)
	}

	e.logger.Info("matching completed",
		zap.String("ride_id", rideID.String()),
		zap.Int("offered_drivers", offered),
	)

	result = "success"
	span.SetAttributes(attribute.String("matching.result", "success"))

	return nil
}

func (e *MatchingEngine) offerCandidate(
	ctx context.Context,
	tx *sql.Tx,
	rideID uuid.UUID,
	attempt int,
	dispatchDecision matching.RetryDecision,
	candidate *candidate.Candidate,
) (bool, error) {

	driverID := candidate.ID

	// Reserve the driver atomically.
	// Another dispatcher may have already reserved this driver.

	resStart := time.Now()
	resCtx, resSpan := matchingTracer.Start(ctx, "dispatch.reservation")
	reserved, err := e.locker.Reserve(resCtx, driverID, rideID)
	resSpan.End()

	resResult := "success"
	if err != nil {
		resResult = "error"
	} else if !reserved {
		resResult = "skipped"
	}
	observability.DispatchReservationDurationSeconds.WithLabelValues(resResult).Observe(time.Since(resStart).Seconds())

	if err != nil {
		e.logger.Error("failed to reserve driver",
			zap.String("ride_id", rideID.String()),
			zap.String("driver_id", driverID.String()),
			zap.Error(err),
		)

		return false, err
	}

	if !reserved {
		observability.DriverOffersTotal.WithLabelValues("reserve_skipped").Inc()
		e.logger.Debug("driver reservation skipped",
			zap.String("ride_id", rideID.String()),
			zap.String("driver_id", driverID.String()),
		)

		return false, nil
	}

	e.logger.Info("driver reserved for offer",
		zap.String("ride_id", rideID.String()),
		zap.String("driver_id", driverID.String()),
		zap.Int("attempt", attempt),
	)

	// Persist the ride offer before publishing any event.
	// If persistence fails the reservation must be released.

	if err := e.driverRepo.InsertRideOfferTx(ctx, tx, rideID, driverID, attempt); err != nil {
		e.logger.Error("failed to insert ride driver offer",
			zap.String("ride_id", rideID.String()),
			zap.String("driver_id", driverID.String()),
			zap.Error(err),
		)
		e.releaseReservedDriverAfterFailure(ctx, rideID, driverID, "insert_offer_failed")

		return false, err
	}

	// -------------------------------------------------------------------------
	// Build the Driver Offered domain event.
	// -------------------------------------------------------------------------

	event := events.DriverOfferedEvent{
		RideID:          rideID,
		DriverID:        driverID,
		OfferTimeoutMs:  dispatchDecision.OfferTimeout.Milliseconds(),
		MatchingAttempt: attempt,
		SearchRadiusKm:  dispatchDecision.RadiusKm,
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
			zap.String("driver_id", driverID.String()),
			zap.Error(err),
		)
		e.releaseReservedDriverAfterFailure(ctx, rideID, driverID, "marshal_offer_event_failed")

		return false, err
	}

	// -------------------------------------------------------------------------
	// Publish the offer asynchronously using the transactional outbox.
	// -------------------------------------------------------------------------

	if err := e.outboxRepo.Insert(ctx, tx,
		outbox.NewEvent(rideID, envelope.Type, payload),
	); err != nil {

		e.logger.Error("failed to insert driver offered outbox event",
			zap.String("ride_id", rideID.String()),
			zap.String("driver_id", driverID.String()),
			zap.Error(err),
		)
		e.releaseReservedDriverAfterFailure(ctx, rideID, driverID, "outbox_insert_failed")

		return false, err
	}

	observability.DriverOffersTotal.WithLabelValues("offered").Inc()

	return true, nil
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
