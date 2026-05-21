package redis

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	appevents "github.com/ashadashraf/ride-hail-app/internal/application/events"
	"github.com/ashadashraf/ride-hail-app/internal/domain/events"
	"github.com/ashadashraf/ride-hail-app/internal/domain/outbox"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/postgres"
	"github.com/ashadashraf/ride-hail-app/internal/ports"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type TimeoutScheduler struct {
	rdb        *redis.Client
	txManager  *postgres.TxManager
	outboxRepo ports.OutboxRepository
}

const redisKey = "driver_timeouts"

func NewTimeoutScheduler(
	rdb *redis.Client,
	txManager *postgres.TxManager,
	outboxRepo ports.OutboxRepository,
) *TimeoutScheduler {
	return &TimeoutScheduler{
		rdb:        rdb,
		txManager:  txManager,
		outboxRepo: outboxRepo,
	}
}

func (s *TimeoutScheduler) Schedule(
	ctx context.Context,
	rideID uuid.UUID,
	driverID uuid.UUID,
	timeout time.Duration,
) error {

	member := rideID.String() + ":" + driverID.String()
	score := float64(time.Now().Add(timeout).Unix())

	rideKey := "ride_timeouts:" + rideID.String()

	pipe := s.rdb.TxPipeline()

	// Add to global timeout ZSET
	pipe.ZAdd(ctx, redisKey, redis.Z{
		Score:  score,
		Member: member,
	})

	// Track driver under this ride (for CancelAll)
	pipe.SAdd(ctx, rideKey, driverID.String())

	// Set TTL to auto cleanup
	pipe.Expire(ctx, rideKey, timeout+10*time.Second)

	_, err := pipe.Exec(ctx)
	return err
}

func (s *TimeoutScheduler) Cancel(
	ctx context.Context,
	rideID uuid.UUID,
	driverID uuid.UUID,
) error {

	member := rideID.String() + ":" + driverID.String()

	return s.rdb.ZRem(ctx, redisKey, member).Err()
}

func (s *TimeoutScheduler) CancelAll(
	ctx context.Context,
	rideID uuid.UUID,
) error {

	key := "ride_timeouts:" + rideID.String()

	driverIDs, err := s.rdb.SMembers(ctx, key).Result()
	if err != nil {
		return err
	}

	if len(driverIDs) == 0 {
		return nil
	}

	var members []string
	for _, d := range driverIDs {
		members = append(members, rideID.String()+":"+d)
	}

	pipe := s.rdb.TxPipeline()

	pipe.ZRem(ctx, redisKey, members)
	pipe.Del(ctx, key)

	_, err = pipe.Exec(ctx)
	return err
}

func (s *TimeoutScheduler) Start(ctx context.Context) {

	ticker := time.NewTicker(1 * time.Second)

	for {
		select {
		case <-ticker.C:
			s.process(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (s *TimeoutScheduler) process(ctx context.Context) {

	now := float64(time.Now().Unix())

	// fetch expired (limit to avoid overload)
	members, err := s.rdb.ZRangeByScore(ctx, redisKey, &redis.ZRangeBy{
		Min:   "0",
		Max:   fmt.Sprintf("%f", now),
		Count: 50,
	}).Result()

	if err != nil {
		return
	}

	for _, m := range members {

		// atomic remove (avoid duplicate processing across workers)
		removed, err := s.rdb.ZRem(ctx, redisKey, m).Result()
		if err != nil || removed == 0 {
			continue
		}

		parts := strings.Split(m, ":")
		if len(parts) != 2 {
			continue
		}

		rideID, _ := uuid.Parse(parts[0])
		driverID, _ := uuid.Parse(parts[1])

		_ = s.emitTimeoutEvent(ctx, rideID, driverID)
	}
}

func (s *TimeoutScheduler) emitTimeoutEvent(
	ctx context.Context,
	rideID uuid.UUID,
	driverID uuid.UUID,
) error {

	return s.txManager.WithinTx(ctx, func(tx *sql.Tx) error {

		event := events.DriverTimeoutEvent{
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

		payload, err := json.Marshal(envelope)
		if err != nil {
			return err
		}

		return s.outboxRepo.Insert(ctx, tx,
			outbox.NewEvent(rideID, envelope.Type, payload),
		)
	})
}
