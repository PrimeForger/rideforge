package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	appevents "github.com/ashadashraf/ride-hail-app/internal/application/events"
	"github.com/ashadashraf/ride-hail-app/internal/domain/events"
	"github.com/ashadashraf/ride-hail-app/internal/domain/outbox"
	"github.com/ashadashraf/ride-hail-app/internal/domain/ride"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/redis"
	"github.com/ashadashraf/ride-hail-app/internal/ports"
	"github.com/google/uuid"
)

type DriverResponseService struct {
	rideRepo   ports.RideRepository
	driverRepo ports.DriverRepository
	locker     ports.DriverLocker
	outboxRepo ports.OutboxRepository
	scheduler  *redis.TimeoutScheduler
}

func NewDriverResponseService(
	rideRepo ports.RideRepository,
	driverRepo ports.DriverRepository,
	locker ports.DriverLocker,
	outboxRepo ports.OutboxRepository,
	scheduler *redis.TimeoutScheduler,
) *DriverResponseService {
	return &DriverResponseService{
		rideRepo:   rideRepo,
		driverRepo: driverRepo,
		locker:     locker,
		outboxRepo: outboxRepo,
		scheduler:  scheduler,
	}
}

func (s *DriverResponseService) HandleDriverAccepted(
	ctx context.Context,
	tx *sql.Tx,
	rideID uuid.UUID,
	driverID uuid.UUID,
) error {

	r, err := s.rideRepo.GetByIDTx(ctx, tx, rideID)
	if err != nil {
		return err
	}

	if err := r.AssignDriver(driverID); err != nil {
		return err
	}

	if err := s.rideRepo.SaveTx(ctx, tx, r); err != nil {
		return err
	}

	if err := s.scheduler.Cancel(ctx, rideID, driverID); err != nil {
		return err
	}

	event := events.RideAcceptedEvent{
		RideID:   rideID,
		DriverID: driverID,
	}

	envelope := appevents.Envelope{
		ID:        uuid.NewString(),
		Type:      event.Name(),
		Aggregate: rideID.String(),
		Data:      event,
		Occurred:  time.Now(),
	}

	payload, _ := json.Marshal(envelope)

	return s.outboxRepo.Insert(ctx, tx,
		outbox.NewEvent(rideID, envelope.Type, payload),
	)
}

func (s *DriverResponseService) HandleDriverRejected(
	ctx context.Context,
	tx *sql.Tx,
	rideID uuid.UUID,
	driverID uuid.UUID,
) error {

	// 1. Mark rejected
	if err := s.driverRepo.MarkDriverRejectedTx(ctx, tx, rideID, driverID); err != nil {
		return err
	}

	// 2. Release driver lock
	if err := s.locker.Release(ctx, driverID); err != nil {
		return err
	}

	// 3. Check ride state
	r, err := s.rideRepo.GetByIDTx(ctx, tx, rideID)
	if err != nil {
		return err
	}

	// If already accepted by another driver → ignore
	if r.Status == ride.StatusAccepted {
		return nil
	}

	// 4. Emit retry event
	event := events.MatchingRetryEvent{
		RideID: rideID,
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
		return err
	}

	return s.outboxRepo.Insert(ctx, tx,
		outbox.NewEvent(rideID, envelope.Type, payload),
	)
}
