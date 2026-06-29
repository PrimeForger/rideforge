package redis

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"
)

type driverLock struct {
	RideID    uuid.UUID `json:"ride_id"`
	CreatedAt int64     `json:"created_at"`
}

type DriverLockerOptions struct {
	LockTTL time.Duration
}

type RedisDriverLocker struct {
	client  *Client
	options DriverLockerOptions
}

func NewRedisDriverLocker(
	client *Client,
	options DriverLockerOptions,
) *RedisDriverLocker {
	return &RedisDriverLocker{
		client:  client,
		options: options,
	}
}

func driverLockKey(driverID uuid.UUID) string {
	return "driver:lock:" + driverID.String()
}

func (l *RedisDriverLocker) Reserve(
	ctx context.Context,
	driverID uuid.UUID,
	rideID uuid.UUID,
) (bool, error) {
	key := driverLockKey(driverID)
	lock := driverLock{
		RideID:    rideID,
		CreatedAt: time.Now().Unix(),
	}

	payload, err := json.Marshal(lock)
	if err != nil {
		return false, err
	}

	res, err := l.client.GetRaw().SetArgs(
		ctx,
		key,
		payload,
		goredis.SetArgs{
			Mode: "NX",
			TTL:  l.options.LockTTL,
		},
	).Result()

	if err != nil {
		return false, err
	}

	return res == "OK", nil
}

var releaseDriverLockScript = goredis.NewScript(`
local current = redis.call("GET", KEYS[1])

if not current then
	return 0
end

local decoded = cjson.decode(current)

if decoded["ride_id"] ~= ARGV[1] then
	return -1
end

redis.call("DEL", KEYS[1])
return 1
`)

func (l *RedisDriverLocker) Release(
	ctx context.Context,
	driverID uuid.UUID,
	rideID uuid.UUID,
) (bool, error) {
	key := driverLockKey(driverID)

	res, err := releaseDriverLockScript.Run(
		ctx,
		l.client.GetRaw(),
		[]string{key},
		rideID.String(),
	).Int()

	if err != nil {
		return false, err
	}

	switch res {
	case 1:
		return true, nil
	case 0:
		return false, nil
	case -1:
		return false, nil
	default:
		return false, errors.New("unexpected redis driver lock release result")
	}
}

func (l *RedisDriverLocker) ForceRelease(
	ctx context.Context,
	driverID uuid.UUID,
) error {
	key := driverLockKey(driverID)

	return l.client.GetRaw().Del(ctx, key).Err()
}
