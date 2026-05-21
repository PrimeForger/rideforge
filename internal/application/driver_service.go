package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	appevents "github.com/ashadashraf/ride-hail-app/internal/application/events"
	"github.com/ashadashraf/ride-hail-app/internal/domain/events"
	"github.com/ashadashraf/ride-hail-app/internal/domain/outbox"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/postgres"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/redis"
	"github.com/ashadashraf/ride-hail-app/internal/ports"
	"github.com/google/uuid"
)

type DriverService struct {
	driverRepo ports.DriverRepository
	locker     ports.DriverLocker
	txManager  *postgres.TxManager
	outboxRepo ports.OutboxRepository
	geo        *redis.GeoService
}

func NewDriverService(
	driverRepo ports.DriverRepository,
	locker ports.DriverLocker,
	txManager *postgres.TxManager,
	outboxRepo ports.OutboxRepository,
	geo *redis.GeoService,
) *DriverService {
	return &DriverService{
		driverRepo: driverRepo,
		locker:     locker,
		txManager:  txManager,
		outboxRepo: outboxRepo,
		geo:        geo,
	}
}

func (s *DriverService) GoOnline(
	ctx context.Context,
	driverID uuid.UUID,
	lat, lng float64,
) error {

	return s.txManager.WithinTx(ctx, func(tx *sql.Tx) error {

		driver, err := s.driverRepo.GetByIDTx(ctx, tx, driverID)
		if err != nil {
			return err
		}

		driver.GoOnline()

		if err := s.driverRepo.SaveTx(ctx, tx, driver); err != nil {
			return err
		}

		// Emit event
		event := events.DriverOnlineEvent{
			DriverID: driverID,
			Lat:      lat,
			Lng:      lng,
		}

		return s.emitEvent(ctx, tx, driverID, event)
	})
}

func (s *DriverService) GoOffline(
	ctx context.Context,
	driverID uuid.UUID,
) error {

	return s.txManager.WithinTx(ctx, func(tx *sql.Tx) error {

		driver, err := s.driverRepo.GetByIDTx(ctx, tx, driverID)
		if err != nil {
			return err
		}

		driver.GoOffline()

		if err := s.driverRepo.SaveTx(ctx, tx, driver); err != nil {
			return err
		}

		event := events.DriverOfflineEvent{
			DriverID: driverID,
		}

		return s.emitEvent(ctx, tx, driverID, event)
	})
}

func (s *DriverService) UpdateLocation(
	ctx context.Context,
	driverID uuid.UUID,
	lat, lng float64,
) error {
	// No need for DB transaction here (hot path)
	return s.geo.UpdateDriverLocation(ctx, driverID, lat, lng)
}

func (s *DriverService) AssignRide(
	ctx context.Context,
	driverID uuid.UUID,
) error {

	return s.txManager.WithinTx(ctx, func(tx *sql.Tx) error {

		driver, err := s.driverRepo.GetByID(ctx, driverID)
		if err != nil {
			return err
		}

		if err := driver.AssignRide(); err != nil {
			return err
		}

		return s.driverRepo.Save(ctx, driver)
	})
}

func (s *DriverService) CompleteRide(
	ctx context.Context,
	driverID uuid.UUID,
) error {

	return s.txManager.WithinTx(ctx, func(tx *sql.Tx) error {

		driver, err := s.driverRepo.GetByID(ctx, driverID)
		if err != nil {
			return err
		}

		driver.CompleteRide()

		if err := s.driverRepo.Save(ctx, driver); err != nil {
			return err
		}

		// TODO: Check wheather this function is complete or not
		// back to geo
		// (location already known, just ensure present)

		return nil
	})
}

func (s *DriverService) emitEvent(
	ctx context.Context,
	tx *sql.Tx,
	driverID uuid.UUID,
	event events.Event,
) error {

	envelope := appevents.Envelope{
		ID:        uuid.NewString(),
		Type:      event.Name(),
		Aggregate: driverID.String(),
		Data:      event,
		Occurred:  time.Now(),
	}

	payload, err := json.Marshal(envelope)
	if err != nil {
		return err
	}

	return s.outboxRepo.Insert(ctx, tx,
		outbox.NewEvent(driverID, envelope.Type, payload),
	)
}
