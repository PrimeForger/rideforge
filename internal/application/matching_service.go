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
	"github.com/ashadashraf/ride-hail-app/internal/ports"
	"github.com/google/uuid"
)

type MatchingService struct {
	driverRepo ports.DriverRepository
	rideSvc    *RideService
	outboxRepo ports.OutboxRepository
}

func NewMatchingService(
	driverRepo ports.DriverRepository,
	rideSvc *RideService,
	outboxRepo ports.OutboxRepository,
) *MatchingService {
	return &MatchingService{
		driverRepo: driverRepo,
		rideSvc:    rideSvc,
		outboxRepo: outboxRepo,
	}
}

// Made for V1 Matching Engine(can be used for testing)
func (m *MatchingService) MatchRide(
	ctx context.Context,
	tx *sql.Tx,
	rideID string,
) error {

	rideUUID, err := uuid.Parse(rideID)
	if err != nil {
		return err
	}

	// 1. Fetch ride INSIDE TX
	r, err := m.rideSvc.rideRepo.GetByIDTx(ctx, tx, rideUUID)
	if err != nil {
		return err
	}

	// 2. Move ride → MATCHING
	if err := r.StartMatching(); err != nil {
		return err
	}

	// 3. Get drivers
	drivers, err := m.driverRepo.GetAvailableDrivers(ctx)
	if err != nil {
		return err
	}

	if len(drivers) == 0 {
		return errors.New("no drivers available")
	}

	driver := drivers[0]

	// 4. Assign driver
	if err := r.AssignDriver(driver.ID); err != nil {
		return err
	}

	// 5. Save ride (with optimistic locking)
	if err := m.rideSvc.rideRepo.SaveTx(ctx, tx, r); err != nil {
		return err
	}

	// 6. Outbox event
	domainEvent := events.RideAcceptedEvent{
		RideID:   r.ID,
		DriverID: driver.ID,
	}

	envelope := appevents.Envelope{
		ID:        uuid.NewString(),
		Type:      domainEvent.Name(),
		Aggregate: r.ID.String(),
		Data:      domainEvent,
		Occurred:  time.Now(),
	}

	payload, err := json.Marshal(envelope)
	if err != nil {
		return err
	}

	outboxEvent := outbox.NewEvent(r.ID, envelope.Type, payload)

	return m.rideSvc.outboxRepo.Insert(ctx, tx, outboxEvent)
}
