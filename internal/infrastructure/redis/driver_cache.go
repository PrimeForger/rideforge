package redis

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/ashadashraf/ride-hail-app/internal/domain/driver"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type DriverCache struct {
	client *Client
}

func NewDriverCache(client *Client) *DriverCache {
	return &DriverCache{client: client}
}

func (c *DriverCache) GetDrivers(ctx context.Context, driverIDs []uuid.UUID) ([]*driver.Driver, error) {

	pipe := c.client.rdb.Pipeline()

	cmds := make([]*redis.MapStringStringCmd, len(driverIDs))

	for i, id := range driverIDs {
		key := "driver:" + id.String()
		cmds[i] = pipe.HGetAll(ctx, key)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return nil, err
	}

	var drivers []*driver.Driver

	for i, cmd := range cmds {

		data := cmd.Val()
		if len(data) == 0 {
			continue
		}

		d, err := parseDriverFromRedis(driverIDs[i].String(), data)
		if err != nil {
			continue
		}

		if d.Rating == 0 {
			d.Rating = 5.0
		}

		if d.Lat == 0 && d.Lng == 0 {
			return nil, errors.New("invalid location")
		}

		drivers = append(drivers, d)
	}

	return drivers, nil
}

func parseDriverFromRedis(driverID string, data map[string]string) (*driver.Driver, error) {

	id, err := uuid.Parse(driverID)
	if err != nil {
		return nil, err
	}

	rating, _ := strconv.ParseFloat(data["rating"], 64)
	acceptance, _ := strconv.ParseFloat(data["acceptance_rate"], 64)
	cancellation, _ := strconv.ParseFloat(data["cancellation_rate"], 64)
	timeout, _ := strconv.ParseFloat(data["timeout_rate"], 64)
	completed, _ := strconv.Atoi(data["completed_rides"])

	var lastAssigned time.Time
	if ts, ok := data["last_assigned_at"]; ok && ts != "" {
		t, err := time.Parse(time.RFC3339, ts)
		if err == nil {
			lastAssigned = t
		}
	}

	lat, _ := strconv.ParseFloat(data["lat"], 64)
	lng, _ := strconv.ParseFloat(data["lng"], 64)

	return &driver.Driver{
		ID:               id,
		Status:           driver.Status(data["status"]),
		Rating:           rating,
		AcceptanceRate:   acceptance,
		CancellationRate: cancellation,
		TimeoutRate:      timeout,
		CompletedRides:   completed,
		LastAssignedAt:   lastAssigned,
		Lat:              lat,
		Lng:              lng,
	}, nil
}

func (c *DriverCache) MarkDriverOffered(
	ctx context.Context,
	rideID uuid.UUID,
	driverID uuid.UUID,
) error {

	key := "ride:" + rideID.String() + ":offered_drivers"

	return c.client.rdb.SAdd(ctx, key, driverID.String()).Err()
}

func (c *DriverCache) IsDriverOffered(
	ctx context.Context,
	rideID uuid.UUID,
	driverID uuid.UUID,
) (bool, error) {

	key := "ride:" + rideID.String() + ":offered_drivers"

	return c.client.rdb.SIsMember(ctx, key, driverID.String()).Result()
}

func (c *DriverCache) ClearOfferedDrivers(
	ctx context.Context,
	rideID uuid.UUID,
) error {

	key := "ride:" + rideID.String() + ":offered_drivers"
	return c.client.rdb.Del(ctx, key).Err()
}

func (c *DriverCache) GetOfferedDrivers(
	ctx context.Context,
	rideID uuid.UUID,
) (map[uuid.UUID]struct{}, error) {

	key := "ride:" + rideID.String() + ":offered_drivers"

	members, err := c.client.rdb.SMembers(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	set := make(map[uuid.UUID]struct{})

	for _, m := range members {
		id, err := uuid.Parse(m)
		if err != nil {
			continue
		}
		set[id] = struct{}{}
	}

	return set, nil
}
