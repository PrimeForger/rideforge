package realtime

import (
	"context"
	"log"

	"github.com/ashadashraf/ride-hail-app/internal/domain/ride"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/observability"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/redis"
	"github.com/ashadashraf/ride-hail-app/internal/ports"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
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

var driverOfferGatewayTracer = otel.Tracer("realtime.driver_offer_gateway")

func (g *DriverOfferGateway) SendOffer(
	ctx context.Context,
	rideID uuid.UUID,
	driverID uuid.UUID,
) error {
	ctx, span := driverOfferGatewayTracer.Start(ctx, "DriverOfferGateway.SendOffer")
	defer span.End()

	span.SetAttributes(
		attribute.String("ride.id", rideID.String()),
		attribute.String("driver.id", driverID.String()),
	)

	connected, err := g.driverCache.IsConnected(ctx, driverID)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.Bool("driver.connected_check_failed", true))
		log.Println("connection state check failed:", err)
	}

	span.SetAttributes(attribute.Bool("driver.connected", connected))

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

			span.SetAttributes(
				attribute.String("delivery.method", "websocket"),
				attribute.String("delivery.status", "delivered_ws"),
			)
			span.SetStatus(codes.Ok, "delivered_ws")

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

		span.SetAttributes(
			attribute.String("delivery.method", "websocket"),
			attribute.String("delivery.status", "ws_failed"),
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

		span.RecordError(err)
		span.SetAttributes(
			attribute.String("delivery.method", "push"),
			attribute.String("delivery.status", "push_failed"),
		)

		// Important: do not return error because timeout safety protects matching.
		span.SetStatus(codes.Ok, "push_failed_timeout_will_protect")

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

	span.SetAttributes(
		attribute.String("delivery.method", "push"),
		attribute.String("delivery.status", "delivered_push"),
	)
	span.SetStatus(codes.Ok, "delivered_push")

	return nil
}
