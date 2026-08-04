package redis

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/ashadashraf/ride-hail-app/internal/domain/driver"
	"github.com/ashadashraf/ride-hail-app/internal/domain/ride"
	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"
)

type DriverCacheOptions struct {
	LocationSeqTTLSeconds int
	HeartbeatTTL          time.Duration
	ConnectionTTL         time.Duration
	DisconnectTTL         time.Duration
	OfferDeliveryTTL      time.Duration
}

type DriverCache struct {
	client  *Client
	options DriverCacheOptions
}

func NewDriverCache(client *Client, options DriverCacheOptions) *DriverCache {
	return &DriverCache{
		client:  client,
		options: options,
	}
}

// Connection

func (c *DriverCache) MarkConnected(
	ctx context.Context,
	driverID uuid.UUID,
) error {
	key := "driver:connection:" + driverID.String()

	return c.client.GetRaw().Set(ctx, key, "online", c.options.ConnectionTTL).Err()
}

func (c *DriverCache) RefreshConnection(
	ctx context.Context,
	driverID uuid.UUID,
) error {
	key := "driver:connection:" + driverID.String()

	return c.client.GetRaw().Expire(ctx, key, c.options.ConnectionTTL).Err()
}

func (c *DriverCache) MarkDisconnected(
	ctx context.Context,
	driverID uuid.UUID,
) error {
	key := "driver:connection:" + driverID.String()

	return c.client.GetRaw().Set(ctx, key, "disconnecting", c.options.DisconnectTTL).Err()
}

func (c *DriverCache) IsConnected(
	ctx context.Context,
	driverID uuid.UUID,
) (bool, error) {
	key := "driver:connection:" + driverID.String()

	val, err := c.client.GetRaw().Get(ctx, key).Result()
	if err != nil {
		if err == goredis.Nil {
			return false, nil
		}
		return false, err
	}

	return val == "online", nil
}

// Connection

func (c *DriverCache) LoadDrivers(ctx context.Context, driverIDs []uuid.UUID) ([]*driver.Driver, error) {

	pipe := c.client.rdb.Pipeline()

	cmds := make([]*goredis.MapStringStringCmd, len(driverIDs))

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

func (c *DriverCache) UpdateDriverLocation(
	ctx context.Context,
	driverID uuid.UUID,
	lat, lng float64,
) error {
	key := "driver:" + driverID.String()

	now := time.Now().Unix()

	return c.client.GetRaw().HSet(ctx, key, map[string]interface{}{
		"lat":        lat,
		"lng":        lng,
		"updated_at": now,
	}).Err()
}

func (c *DriverCache) UpdateDriverLocationDetails(
	ctx context.Context,
	driverID uuid.UUID,
	lat, lng float64,
	accuracy float64,
	speed float64,
	bearing float64,
	seq int64,
) error {
	key := "driver:" + driverID.String()

	now := time.Now().Unix()

	return c.client.GetRaw().HSet(ctx, key, map[string]interface{}{
		"lat":        lat,
		"lng":        lng,
		"accuracy":   accuracy,
		"speed":      speed,
		"bearing":    bearing,
		"seq":        seq,
		"updated_at": now,
	}).Err()
}

func (c *DriverCache) MarkOfferAcked(
	ctx context.Context,
	rideID uuid.UUID,
	driverID uuid.UUID,
) error {
	key := "ride_offer_ack:" + rideID.String() + ":" + driverID.String()

	return c.client.GetRaw().Set(
		ctx,
		key,
		time.Now().Unix(),
		30*time.Minute,
	).Err()
}

func (c *DriverCache) IsOfferAcked(
	ctx context.Context,
	rideID uuid.UUID,
	driverID uuid.UUID,
) (bool, error) {
	key := "ride_offer_ack:" + rideID.String() + ":" + driverID.String()

	exists, err := c.client.GetRaw().Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}

	return exists == 1, nil
}

func (c *DriverCache) MarkOnline(
	ctx context.Context,
	driverID uuid.UUID,
	lat, lng float64,
) error {
	key := "driver:" + driverID.String()

	pipe := c.client.GetRaw().TxPipeline()

	pipe.HSet(ctx, key, map[string]interface{}{
		"status":     "ONLINE",
		"lat":        lat,
		"lng":        lng,
		"updated_at": time.Now().Unix(),
	})

	pipe.SAdd(ctx, "drivers:online", driverID.String())

	_, err := pipe.Exec(ctx)
	return err
}

func (c *DriverCache) MarkOffline(
	ctx context.Context,
	driverID uuid.UUID,
) error {
	key := "driver:" + driverID.String()

	pipe := c.client.GetRaw().TxPipeline()

	pipe.HSet(ctx, key, map[string]interface{}{
		"status":     "OFFLINE",
		"updated_at": time.Now().Unix(),
	})

	pipe.SRem(ctx, "drivers:online", driverID.String())

	_, err := pipe.Exec(ctx)
	return err
}

func (c *DriverCache) RefreshHeartbeat(
	ctx context.Context,
	driverID uuid.UUID,
) error {
	key := "driver:heartbeat:" + driverID.String()

	return c.client.GetRaw().Set(ctx, key, "1", c.options.HeartbeatTTL).Err()
}

func (c *DriverCache) GetOnlineDriverIDs(
	ctx context.Context,
) ([]uuid.UUID, error) {
	members, err := c.client.GetRaw().SMembers(ctx, "drivers:online").Result()
	if err != nil {
		return nil, err
	}

	ids := make([]uuid.UUID, 0, len(members))

	for _, raw := range members {
		id, err := uuid.Parse(raw)
		if err != nil {
			continue
		}
		ids = append(ids, id)
	}

	return ids, nil
}

func (c *DriverCache) HasHeartbeat(
	ctx context.Context,
	driverID uuid.UUID,
) (bool, error) {
	key := "driver:heartbeat:" + driverID.String()

	exists, err := c.client.GetRaw().Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}

	return exists == 1, nil
}

var acceptLocationSeqScript = goredis.NewScript(`
local current = redis.call("GET", KEYS[1])

if current and tonumber(ARGV[1]) <= tonumber(current) then
	return 0
end

redis.call("SET", KEYS[1], ARGV[1], "EX", ARGV[2])
return 1
`)

func (c *DriverCache) AcceptLocationSeq(
	ctx context.Context,
	driverID uuid.UUID,
	seq int64,
) (bool, error) {
	if seq <= 0 {
		return false, nil
	}

	key := "driver:last_location_seq:" + driverID.String()

	res, err := acceptLocationSeqScript.Run(
		ctx,
		c.client.GetRaw(),
		[]string{key},
		seq,
		c.options.LocationSeqTTLSeconds,
	).Int()

	if err != nil {
		return false, err
	}

	return res == 1, nil
}

func offerDeliveryKey(rideID uuid.UUID, driverID uuid.UUID) string {
	return "ride_offer_delivery:" + rideID.String() + ":" + driverID.String()
}

func (c *DriverCache) MarkOfferDeliveryStatus(
	ctx context.Context,
	rideID uuid.UUID,
	driverID uuid.UUID,
	status ride.OfferDeliveryStatus,
) error {
	key := offerDeliveryKey(rideID, driverID)

	pipe := c.client.GetRaw().TxPipeline()

	pipe.HSet(ctx, key, map[string]interface{}{
		"status":     string(status),
		"updated_at": time.Now().Unix(),
	})

	pipe.Expire(ctx, key, c.options.OfferDeliveryTTL)

	_, err := pipe.Exec(ctx)
	return err
}

func (c *DriverCache) GetOfferDeliveryStatus(
	ctx context.Context,
	rideID uuid.UUID,
	driverID uuid.UUID,
) (ride.OfferDeliveryStatus, error) {
	key := offerDeliveryKey(rideID, driverID)

	status, err := c.client.GetRaw().HGet(ctx, key, "status").Result()
	if err != nil {
		if err == goredis.Nil {
			return "", nil
		}
		return "", err
	}

	return ride.OfferDeliveryStatus(status), nil
}

func driverPushTokensKey(driverID uuid.UUID) string {
	return "driver:push_tokens:" + driverID.String()
}

func (c *DriverCache) AddPushToken(
	ctx context.Context,
	driverID uuid.UUID,
	token string,
) error {
	key := driverPushTokensKey(driverID)

	return c.client.GetRaw().SAdd(ctx, key, token).Err()
}

func (c *DriverCache) GetPushTokens(
	ctx context.Context,
	driverID uuid.UUID,
) ([]string, error) {
	key := driverPushTokensKey(driverID)

	return c.client.GetRaw().SMembers(ctx, key).Result()
}

func (c *DriverCache) UpdateDriverMetrics(
	ctx context.Context,
	driverID uuid.UUID,
	acceptanceRate float64,
	cancellationRate float64,
	timeoutRate float64,
	completedRides int64,
) error {
	key := "driver:" + driverID.String()

	return c.client.GetRaw().HSet(ctx, key, map[string]interface{}{
		"acceptance_rate":    acceptanceRate,
		"cancellation_rate":  cancellationRate,
		"timeout_rate":       timeoutRate,
		"completed_rides":    completedRides,
		"metrics_updated_at": time.Now().Unix(),
	}).Err()
}
