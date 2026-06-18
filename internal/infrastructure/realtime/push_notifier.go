package realtime

import (
	"context"
	"log"

	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/redis"
	"github.com/google/uuid"
)

type NoopPushNotifier struct {
	driverCache *redis.DriverCache
}

func NewNoopPushNotifier(driverCache *redis.DriverCache) *NoopPushNotifier {
	return &NoopPushNotifier{driverCache: driverCache}
}

func (n *NoopPushNotifier) SendRideOffer(
	ctx context.Context,
	rideID uuid.UUID,
	driverID uuid.UUID,
) error {
	tokens, err := n.driverCache.GetPushTokens(ctx, driverID)
	if err != nil {
		return err
	}

	if len(tokens) == 0 {
		log.Printf("no push tokens found: driver_id=%s", driverID)
		return nil
	}

	log.Printf("would send push to %d tokens: ride_id=%s driver_id=%s", len(tokens), rideID, driverID)
	return nil
}
