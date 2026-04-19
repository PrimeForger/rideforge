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

type MatchingEngine struct {
	driverRepo ports.DriverRepository
	locker     ports.DriverLocker
	outboxRepo ports.OutboxRepository
}

func NewMatchingEngine(
	driverRepo ports.DriverRepository,
	locker ports.DriverLocker,
	outboxRepo ports.OutboxRepository,
) *MatchingEngine {
	return &MatchingEngine{
		driverRepo: driverRepo,
		locker:     locker,
		outboxRepo: outboxRepo,
	}
}

const maxDriverAttempts = 5

func (e *MatchingEngine) HandleMatchingStarted(
	ctx context.Context,
	tx *sql.Tx,
	rideID uuid.UUID,
) error {

	// Count attempts
	attemptCount, err := e.driverRepo.CountRideAttemptsTx(ctx, tx, rideID)
	if err != nil {
		return err
	}

	if attemptCount >= maxDriverAttempts {
		return errors.New("max driver attempts reached")
	}

	// 2. Get available drivers excluding already tried
	drivers, err := e.driverRepo.GetAvailableDriversExcludingTx(ctx, tx, rideID)
	if err != nil {
		return err
	}

	for _, d := range drivers {

		ok, err := e.locker.Reserve(ctx, d.ID, rideID)
		if err != nil {
			return err
		}

		if !ok {
			continue
		}

		// Record attempt
		err = e.driverRepo.InsertRideOfferTx(ctx, tx, rideID, d.ID, attemptCount+1)
		if err != nil {
			return err
		}

		event := events.DriverOfferedEvent{
			RideID:   rideID,
			DriverID: d.ID,
		}

		envelope := appevents.Envelope{
			ID:        uuid.NewString(),
			Type:      event.Name(),
			Aggregate: rideID.String(),
			Data:      event,
			Occurred:  time.Now(),
		}

		payload, _ := json.Marshal(envelope)

		return e.outboxRepo.Insert(ctx, tx,
			outbox.NewEvent(rideID, envelope.Type, payload),
		)
	}

	return errors.New("no drivers available")
}
