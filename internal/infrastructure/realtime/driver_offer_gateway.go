package realtime

import (
	"context"
	"log"

	"github.com/ashadashraf/ride-hail-app/internal/domain/ride"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/observability"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/redis"
	"github.com/ashadashraf/ride-hail-app/internal/ports"
	"github.com/google/uuid"
)

type DriverOfferGateway struct {
	hub         *Hub
	driverCache *redis.DriverCache
	push        ports.PushNotifier
}

func NewDriverOfferGateway(
	hub *Hub,
	driverCache *redis.DriverCache,
	push ports.PushNotifier,
) *DriverOfferGateway {
	return &DriverOfferGateway{
		hub:         hub,
		driverCache: driverCache,
		push:        push,
	}
}

func (g *DriverOfferGateway) SendOffer(
	ctx context.Context,
	rideID uuid.UUID,
	driverID uuid.UUID,
) error {
	connected, err := g.driverCache.IsConnected(ctx, driverID)
	if err != nil {
		log.Println("connection state check failed:", err)
	}

	if connected {
		ok := g.hub.SendToDriver(ctx, driverID, OutgoingMessage{
			Type: "ride.offer.received",
			Data: map[string]string{
				"ride_id":      rideID.String(),
				"driver_id":    driverID.String(),
				"ack_required": "true",
			},
		})

		if ok {
			observability.DriverOffersTotal.WithLabelValues("delivered_ws").Inc()

			_ = g.driverCache.MarkOfferDeliveryStatus(
				ctx,
				rideID,
				driverID,
				ride.OfferDeliveryWebSocketSent,
			)

			log.Printf("ride offer delivered over websocket: ride_id=%s driver_id=%s", rideID, driverID)
			return nil
		}

		observability.DriverOffersTotal.WithLabelValues("ws_failed").Inc()

		_ = g.driverCache.MarkOfferDeliveryStatus(
			ctx,
			rideID,
			driverID,
			ride.OfferDeliveryWsFailed,
		)
		log.Printf("websocket send failed, falling back to push: ride_id=%s driver_id=%s", rideID, driverID)
	}

	if err := g.push.SendRideOffer(ctx, rideID, driverID); err != nil {
		observability.DriverOffersTotal.WithLabelValues("push_failed").Inc()

		_ = g.driverCache.MarkOfferDeliveryStatus(
			ctx,
			rideID,
			driverID,
			ride.OfferDeliveryPushFailed,
		)

		log.Printf("push fallback failed: ride_id=%s driver_id=%s err=%v", rideID, driverID, err)
		return nil
	}

	observability.DriverOffersTotal.WithLabelValues("delivered_push").Inc()

	_ = g.driverCache.MarkOfferDeliveryStatus(
		ctx,
		rideID,
		driverID,
		ride.OfferDeliveryPushSent,
	)
	return nil
}
