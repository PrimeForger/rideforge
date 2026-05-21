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
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/redis"
	"github.com/ashadashraf/ride-hail-app/internal/ports"
	"github.com/google/uuid"
)

type MatchingEngine struct {
	driverRepo  ports.DriverRepository
	locker      ports.DriverLocker
	outboxRepo  ports.OutboxRepository
	geo         *redis.GeoService
	driverCache *redis.DriverCache
	ranking     matching.Ranker
	cfg         *config.Config
}

func NewMatchingEngine(
	driverRepo ports.DriverRepository,
	locker ports.DriverLocker,
	outboxRepo ports.OutboxRepository,
	geo *redis.GeoService,
	driverCache *redis.DriverCache,
	ranking matching.Ranker,
	cfg *config.Config,
) *MatchingEngine {
	return &MatchingEngine{
		driverRepo:  driverRepo,
		locker:      locker,
		outboxRepo:  outboxRepo,
		geo:         geo,
		driverCache: driverCache,
		ranking:     ranking,
		cfg:         cfg,
	}
}

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

	if attemptCount >= e.cfg.MaxDriverAttempts {
		return errors.New("max driver attempts reached")
	}

	// Dynamic radius expansion
	radius := e.computeRadius(attemptCount)

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
		return err
	}

	if len(nearby) == 0 {
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
		return err
	}

	if len(drivers) == 0 {
		return errors.New("no eligible drivers")
	}

	offeredSet, err := e.driverCache.GetOfferedDrivers(ctx, rideID)
	if err != nil {
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
		return errors.New("no eligible drivers after filtering")
	}

	// Offer top N drivers (parallel batch)
	selected := 0

	for h.Len() > 0 && selected < e.cfg.OfferBatchSize {

		candidate := heap.Pop(h).(matching.Candidate)

		ok, err := e.locker.ReserveTx(ctx, tx, candidate.DriverID, rideID)
		if err != nil {
			return err
		}

		if !ok {
			continue
		}

		// Record attempt in DB
		err = e.driverRepo.InsertRideOfferTx(ctx, tx, rideID, candidate.DriverID, attemptCount+1)
		if err != nil {
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

		payload, _ := json.Marshal(envelope)

		if err := e.outboxRepo.Insert(ctx, tx,
			outbox.NewEvent(rideID, envelope.Type, payload),
		); err != nil {
			return err
		}

		selected++
	}

	if selected == 0 {
		return errors.New("no drivers reserved")
	}

	return nil
}

func (e *MatchingEngine) computeRadius(attempt int) float64 {
	return e.cfg.SearchRadiusKm * math.Pow(2, float64(attempt))
}
