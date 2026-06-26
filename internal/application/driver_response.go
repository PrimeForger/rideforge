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

func (s *DriverResponseService) HandleDriverAccepted(
	ctx context.Context,
	tx *sql.Tx,
	rideID uuid.UUID,
	driverID uuid.UUID,
) error {

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
		return err
	}

	// Already accepted by another driver.
	if r.Status == ride.StatusAccepted {
		s.logger.Warn("driver acceptance ignored because ride already accepted",
			zap.String("ride_id", rideID.String()),
			zap.String("driver_id", driverID.String()),
		)
		return nil
	}

	if r.Status != ride.StatusMatching {
		s.logger.Warn("driver acceptance rejected because ride is not matching",
			zap.String("ride_id", rideID.String()),
			zap.String("driver_id", driverID.String()),
			zap.String("ride_status", string(r.Status)),
		)
		return errors.New("ride is not in matching state")
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
		return err
	}

	// Get other active offered drivers before expiring them.
	otherDrivers, err := s.driverRepo.GetActiveOfferDriversTx(ctx, tx, rideID, driverID)
	if err != nil {
		return err
	}

	if err := r.AssignDriver(driverID); err != nil {
		return err
	}

	if err := s.rideRepo.SaveTx(ctx, tx, r); err != nil {
		return err
	}

	// Winner becomes BUSY.
	if err := s.driverRepo.MarkDriverBusyTx(ctx, tx, driverID); err != nil {
		return err
	}

	// Expire loser offers.
	if err := s.driverRepo.ExpireOtherOffersTx(ctx, tx, rideID, driverID); err != nil {
		return err
	}

	// Release other reserved drivers.
	for _, otherDriverID := range otherDrivers {
		ok, err := s.locker.ReleaseTx(ctx, tx, otherDriverID, rideID)
		if err != nil {
			return err
		}

		if !ok {
			s.logger.Warn("loser driver release skipped",
				zap.String("ride_id", rideID.String()),
				zap.String("driver_id", otherDriverID.String()),
			)
		}
	}

	event := events.RideAcceptedEvent{
		RideID:   rideID,
		DriverID: driverID,
	}

	if err := s.insertOutboxEvent(ctx, tx, rideID, event); err != nil {
		return err
	}

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

	s.logger.Info("driver rejected ride",
		zap.String("ride_id", rideID.String()),
		zap.String("driver_id", driverID.String()),
	)

	// Check ride state
	r, err := s.rideRepo.GetByIDTx(ctx, tx, rideID)
	if err != nil {
		return err
	}

	// If already accepted by another driver → ignore
	if r.Status == ride.StatusAccepted {
		s.logger.Warn("driver rejection ignored because ride already accepted",
			zap.String("ride_id", rideID.String()),
			zap.String("driver_id", driverID.String()),
		)
		return nil
	}

	// Mark rejected
	if err := s.driverRepo.MarkDriverRejectedTx(ctx, tx, rideID, driverID); err != nil {
		return err
	}

	// Release driver lock
	if ok, err := s.locker.ReleaseTx(ctx, tx, driverID, rideID); err != nil {
		return err
	} else if !ok {
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

	return s.insertOutboxEvent(ctx, tx, rideID, event)
}

func (s *DriverResponseService) HandleDriverTimeout(
	ctx context.Context,
	tx *sql.Tx,
	rideID uuid.UUID,
	driverID uuid.UUID,
	offerAcked bool,
	deliveryStatus string,
) error {

	s.logger.Warn("driver offer timed out",
		zap.String("ride_id", rideID.String()),
		zap.String("driver_id", driverID.String()),
		zap.Bool("offer_acked", offerAcked),
		zap.String("delivery_status", deliveryStatus),
	)

	// 1. Check ride state (important!)
	r, err := s.rideRepo.GetByIDTx(ctx, tx, rideID)
	if err != nil {
		return err
	}

	// If already accepted → ignore timeout
	if r.Status == ride.StatusAccepted {
		s.logger.Warn("driver timeout ignored because ride already accepted",
			zap.String("ride_id", rideID.String()),
			zap.String("driver_id", driverID.String()),
		)
		return nil
	}

	// 2. Mark offer as TIMEOUT (or REJECTED)
	if err := s.driverRepo.MarkDriverTimeoutTx(ctx, tx, rideID, driverID); err != nil {
		return err
	}

	// 3. Release lock
	if ok, err := s.locker.ReleaseTx(ctx, tx, driverID, rideID); err != nil {
		return err
	} else if !ok {
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

	processedEvent := events.DriverTimeoutProcessedEvent{
		RideID:         rideID,
		DriverID:       driverID,
		OfferAcked:     offerAcked,
		DeliveryStatus: deliveryStatus,
		TimeoutReason:  reason,
	}

	if err := s.insertOutboxEvent(ctx, tx, rideID, processedEvent); err != nil {
		return err
	}

	// 4. Emit retry event
	retryEvent := events.MatchingRetryEvent{
		RideID: rideID,
	}

	if err := s.insertOutboxEvent(ctx, tx, rideID, retryEvent); err != nil {
		return err
	}

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
