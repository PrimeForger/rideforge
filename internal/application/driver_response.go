package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	appevents "github.com/ashadashraf/ride-hail-app/internal/application/events"
	"github.com/ashadashraf/ride-hail-app/internal/domain/events"
	"github.com/ashadashraf/ride-hail-app/internal/domain/outbox"
	"github.com/ashadashraf/ride-hail-app/internal/domain/ride"
	"github.com/ashadashraf/ride-hail-app/internal/ports"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

type DriverResponseService struct {
	rideRepo   ports.RideRepository
	driverRepo ports.DriverRepository
	locker     ports.DriverLocker
	outboxRepo ports.OutboxRepository
	logger     *zap.Logger
}

func NewDriverResponseService(
	rideRepo ports.RideRepository,
	driverRepo ports.DriverRepository,
	locker ports.DriverLocker,
	outboxRepo ports.OutboxRepository,
	logger *zap.Logger,
) *DriverResponseService {
	return &DriverResponseService{
		rideRepo:   rideRepo,
		driverRepo: driverRepo,
		locker:     locker,
		outboxRepo: outboxRepo,
		logger:     logger,
	}
}

var driverResponseTracer = otel.Tracer("application.driver_response")

func (s *DriverResponseService) HandleDriverAccepted(
	ctx context.Context,
	tx *sql.Tx,
	rideID uuid.UUID,
	driverID uuid.UUID,
) error {
	ctx, span := driverResponseTracer.Start(ctx, "DriverResponseService.HandleDriverAccepted")
	defer span.End()

	span.SetAttributes(
		attribute.String("ride.id", rideID.String()),
		attribute.String("driver.id", driverID.String()),
	)

	fail := func(err error, status string) error {
		span.RecordError(err)
		span.SetAttributes(attribute.String("driver_response.result", status))
		span.SetStatus(codes.Error, status)
		return err
	}

	s.logger.Info("driver accepted ride",
		zap.String("ride_id", rideID.String()),
		zap.String("driver_id", driverID.String()),
	)

	r, err := s.rideRepo.GetByIDTx(ctx, tx, rideID)
	if err != nil {
		s.logger.Error("failed to fetch ride for driver acceptance",
			zap.String("ride_id", rideID.String()),
			zap.String("driver_id", driverID.String()),
			zap.Error(err),
		)
		return fail(err, "ride_fetch_failed")
	}

	span.SetAttributes(attribute.String("ride.status", string(r.Status)))

	// Already accepted by another driver.
	if r.Status == ride.StatusAccepted {
		span.SetAttributes(attribute.String("driver_response.result", "already_accepted"))
		span.SetStatus(codes.Ok, "already accepted")

		s.logger.Warn("driver acceptance ignored because ride already accepted",
			zap.String("ride_id", rideID.String()),
			zap.String("driver_id", driverID.String()),
		)
		return nil
	}

	if r.Status != ride.StatusMatching {
		err := errors.New("ride is not in matching state")

		s.logger.Warn("driver acceptance rejected because ride is not matching",
			zap.String("ride_id", rideID.String()),
			zap.String("driver_id", driverID.String()),
			zap.String("ride_status", string(r.Status)),
		)
		return fail(err, "ride_not_matching")
	}

	// Mark this offer accepted.
	if err := s.driverRepo.MarkDriverAcceptedTx(
		ctx,
		tx,
		rideID,
		driverID,
	); err != nil {
		s.logger.Error("failed to mark driver offer accepted",
			zap.String("ride_id", rideID.String()),
			zap.String("driver_id", driverID.String()),
			zap.Error(err),
		)
		return fail(err, "mark_offer_accepted_failed")
	}

	// Get other active offered drivers before expiring them.
	otherDrivers, err := s.driverRepo.GetActiveOfferDriversTx(ctx, tx, rideID, driverID)
	if err != nil {
		return fail(err, "active_offer_fetch_failed")
	}

	span.SetAttributes(attribute.Int("driver_response.other_offer_count", len(otherDrivers)))

	if err := r.AssignDriver(driverID); err != nil {
		return fail(err, "assign_driver_failed")
	}

	if err := s.rideRepo.SaveTx(ctx, tx, r); err != nil {
		return fail(err, "ride_save_failed")
	}

	// Winner becomes BUSY.
	if err := s.driverRepo.MarkDriverBusyTx(ctx, tx, driverID); err != nil {
		return fail(err, "mark_driver_busy_failed")
	}

	// Expire loser offers.
	if err := s.driverRepo.ExpireOtherOffersTx(ctx, tx, rideID, driverID); err != nil {
		return fail(err, "expire_other_offers_failed")
	}

	releasedOtherDrivers := 0

	// Release other reserved drivers.
	for _, otherDriverID := range otherDrivers {
		ok, err := s.locker.ReleaseTx(ctx, tx, otherDriverID, rideID)
		if err != nil {
			return fail(err, "release_loser_driver_failed")
		}

		if !ok {
			s.logger.Warn("loser driver release skipped",
				zap.String("ride_id", rideID.String()),
				zap.String("driver_id", otherDriverID.String()),
			)
			continue
		}

		releasedOtherDrivers++
	}

	event := events.RideAcceptedEvent{
		RideID:   rideID,
		DriverID: driverID,
	}

	if err := s.insertOutboxEvent(ctx, tx, rideID, event); err != nil {
		return fail(err, "ride_accepted_outbox_failed")
	}

	span.SetAttributes(
		attribute.String("driver_response.result", "accepted"),
		attribute.Int("driver_response.expired_other_offers", len(otherDrivers)),
		attribute.Int("driver_response.released_other_drivers", releasedOtherDrivers),
	)
	span.SetStatus(codes.Ok, "accepted")

	s.logger.Info("ride accepted successfully",
		zap.String("ride_id", rideID.String()),
		zap.String("driver_id", driverID.String()),
		zap.Int("expired_other_offers", len(otherDrivers)),
	)

	return nil
}

func (s *DriverResponseService) HandleDriverRejected(
	ctx context.Context,
	tx *sql.Tx,
	rideID uuid.UUID,
	driverID uuid.UUID,
) error {
	ctx, span := driverResponseTracer.Start(ctx, "DriverResponseService.HandleDriverRejected")
	defer span.End()

	span.SetAttributes(
		attribute.String("ride.id", rideID.String()),
		attribute.String("driver.id", driverID.String()),
	)

	fail := func(err error, status string) error {
		span.RecordError(err)
		span.SetAttributes(attribute.String("driver_response.result", status))
		span.SetStatus(codes.Error, status)
		return err
	}

	s.logger.Info("driver rejected ride",
		zap.String("ride_id", rideID.String()),
		zap.String("driver_id", driverID.String()),
	)

	// Check ride state
	r, err := s.rideRepo.GetByIDTx(ctx, tx, rideID)
	if err != nil {
		return fail(err, "ride_fetch_failed")
	}

	span.SetAttributes(attribute.String("ride.status", string(r.Status)))

	// If already accepted by another driver → ignore
	if r.Status == ride.StatusAccepted {
		span.SetAttributes(attribute.String("driver_response.result", "already_accepted"))
		span.SetStatus(codes.Ok, "already accepted")

		s.logger.Warn("driver rejection ignored because ride already accepted",
			zap.String("ride_id", rideID.String()),
			zap.String("driver_id", driverID.String()),
		)
		return nil
	}

	// Mark rejected
	if err := s.driverRepo.MarkDriverRejectedTx(ctx, tx, rideID, driverID); err != nil {
		return fail(err, "mark_offer_rejected_failed")
	}

	// Release driver lock
	if ok, err := s.locker.ReleaseTx(ctx, tx, driverID, rideID); err != nil {
		return fail(err, "release_driver_failed")
	} else if !ok {
		span.SetAttributes(attribute.String("driver_response.result", "release_skipped"))
		span.SetStatus(codes.Ok, "release skipped")

		s.logger.Warn("release skipped: driver not reserved for ride",
			zap.String("ride_id", rideID.String()),
			zap.String("driver_id", driverID.String()),
		)
		return nil
	}

	// Emit processed event (NOT retry)
	event := events.DriverRejectedProcessedEvent{
		RideID:   rideID,
		DriverID: driverID,
	}

	if err := s.insertOutboxEvent(ctx, tx, rideID, event); err != nil {
		return fail(err, "driver_rejected_processed_outbox_failed")
	}

	span.SetAttributes(attribute.String("driver_response.result", "rejected"))
	span.SetStatus(codes.Ok, "rejected")

	return nil
}

func (s *DriverResponseService) HandleDriverTimeout(
	ctx context.Context,
	tx *sql.Tx,
	rideID uuid.UUID,
	driverID uuid.UUID,
	offerAcked bool,
	deliveryStatus string,
) error {
	ctx, span := driverResponseTracer.Start(ctx, "DriverResponseService.HandleDriverTimeout")
	defer span.End()

	span.SetAttributes(
		attribute.String("ride.id", rideID.String()),
		attribute.String("driver.id", driverID.String()),
		attribute.Bool("offer.acked", offerAcked),
		attribute.String("offer.delivery_status", deliveryStatus),
	)

	fail := func(err error, status string) error {
		span.RecordError(err)
		span.SetAttributes(attribute.String("driver_response.result", status))
		span.SetStatus(codes.Error, status)
		return err
	}

	s.logger.Warn("driver offer timed out",
		zap.String("ride_id", rideID.String()),
		zap.String("driver_id", driverID.String()),
		zap.Bool("offer_acked", offerAcked),
		zap.String("delivery_status", deliveryStatus),
	)

	// 1. Check ride state (important!)
	r, err := s.rideRepo.GetByIDTx(ctx, tx, rideID)
	if err != nil {
		return fail(err, "ride_fetch_failed")
	}

	span.SetAttributes(attribute.String("ride.status", string(r.Status)))

	// If already accepted → ignore timeout
	if r.Status == ride.StatusAccepted {
		span.SetAttributes(attribute.String("driver_response.result", "already_accepted"))
		span.SetStatus(codes.Ok, "already accepted")

		s.logger.Warn("driver timeout ignored because ride already accepted",
			zap.String("ride_id", rideID.String()),
			zap.String("driver_id", driverID.String()),
		)
		return nil
	}

	// 2. Mark offer as TIMEOUT (or REJECTED)
	if err := s.driverRepo.MarkDriverTimeoutTx(ctx, tx, rideID, driverID); err != nil {
		return fail(err, "mark_offer_timeout_failed")
	}

	// 3. Release lock
	if ok, err := s.locker.ReleaseTx(ctx, tx, driverID, rideID); err != nil {
		return fail(err, "release_driver_failed")
	} else if !ok {
		span.SetAttributes(attribute.String("driver_response.result", "release_skipped"))
		span.SetStatus(codes.Ok, "release skipped")

		s.logger.Warn("release skipped: driver not reserved for ride",
			zap.String("ride_id", rideID.String()),
			zap.String("driver_id", driverID.String()),
		)
		return nil
	}

	reason := "DELIVERY_UNCONFIRMED"
	if offerAcked {
		reason = "ACKED_BUT_NO_RESPONSE"
	} else if deliveryStatus == string(ride.OfferDeliveryWebSocketSent) ||
		deliveryStatus == string(ride.OfferDeliveryPushSent) {
		reason = "DELIVERED_BUT_NOT_ACKED"
	}

	span.SetAttributes(attribute.String("offer.timeout_reason", reason))

	processedEvent := events.DriverTimeoutProcessedEvent{
		RideID:         rideID,
		DriverID:       driverID,
		OfferAcked:     offerAcked,
		DeliveryStatus: deliveryStatus,
		TimeoutReason:  reason,
	}

	if err := s.insertOutboxEvent(ctx, tx, rideID, processedEvent); err != nil {
		return fail(err, "driver_timeout_processed_outbox_failed")
	}

	// 4. Emit retry event
	retryEvent := events.MatchingRetryEvent{
		RideID: rideID,
	}

	if err := s.insertOutboxEvent(ctx, tx, rideID, retryEvent); err != nil {
		return fail(err, "matching_retry_outbox_failed")
	}

	span.SetAttributes(attribute.String("driver_response.result", "timeout_processed"))
	span.SetStatus(codes.Ok, "timeout processed")

	s.logger.Info("matching retry emitted after driver timeout",
		zap.String("ride_id", rideID.String()),
		zap.String("driver_id", driverID.String()),
		zap.String("timeout_reason", reason),
	)

	return nil
}

func (s *DriverResponseService) insertOutboxEvent(
	ctx context.Context,
	tx *sql.Tx,
	aggregateID uuid.UUID,
	event events.Event,
) error {
	envelope := appevents.Envelope{
		ID:        uuid.NewString(),
		Type:      event.Name(),
		Aggregate: aggregateID.String(),
		Data:      event,
		Occurred:  time.Now(),
	}

	payload, err := json.Marshal(envelope)
	if err != nil {
		return err
	}

	return s.outboxRepo.Insert(ctx, tx,
		outbox.NewEvent(aggregateID, envelope.Type, payload),
	)
}
