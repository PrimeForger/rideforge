package realtime

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/ashadashraf/ride-hail-app/internal/application"
	"github.com/ashadashraf/ride-hail-app/internal/config"
	"github.com/ashadashraf/ride-hail-app/internal/domain/ride"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/redis"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 30 * time.Second
	pingPeriod     = 20 * time.Second
	maxMessageSize = 1024
	sendBufferSize = 256
)

type Client struct {
	driverID uuid.UUID
	conn     *websocket.Conn
	hub      *Hub

	geo         *redis.GeoService
	driverCache *redis.DriverCache

	driverLocationService *application.DriverLocationService

	send chan OutgoingMessage
	done chan struct{}

	maxAccuracyMeters float64
	minLocationGap    time.Duration
	lastLocationAt    time.Time

	cfg *config.RealtimeConfig
}

func NewClient(
	driverID uuid.UUID,
	conn *websocket.Conn,
	hub *Hub,
	geo *redis.GeoService,
	driverCache *redis.DriverCache,
	driverLocationService *application.DriverLocationService,
	cfg *config.RealtimeConfig,
) *Client {
	return &Client{
		driverID:              driverID,
		conn:                  conn,
		hub:                   hub,
		geo:                   geo,
		driverCache:           driverCache,
		driverLocationService: driverLocationService,
		send:                  make(chan OutgoingMessage, sendBufferSize),
		done:                  make(chan struct{}),
		cfg:                   cfg,
		maxAccuracyMeters:     cfg.MaxLocationAccuracyMeters,
		minLocationGap:        time.Duration(cfg.MinLocationIntervalMs) * time.Millisecond,
	}
}

func (c *Client) Start(ctx context.Context) {
	c.hub.Register(c.driverID, c)

	if err := c.driverCache.MarkConnected(ctx, c.driverID); err != nil {
		log.Println("failed to mark driver connected:", err)
	}

	go c.writePump(ctx)
	c.readPump(ctx)
}

func (c *Client) Close() {
	select {
	case <-c.done:
		return
	default:
		close(c.done)
		_ = c.conn.Close()
	}
}

func (c *Client) Send(ctx context.Context, msg OutgoingMessage) bool {
	select {
	case c.send <- msg:
		return true
	case <-ctx.Done():
		return false
	case <-c.done:
		return false
	default:
		log.Println("driver websocket send buffer full:", c.driverID)
		return false
	}
}

func (c *Client) readPump(ctx context.Context) {
	defer func() {
		c.hub.Unregister(c.driverID, c)

		if err := c.driverCache.MarkDisconnected(ctx, c.driverID); err != nil {
			log.Println("failed to mark driver disconnected:", err)
		}

		c.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		var msg IncomingMessage

		if err := c.conn.ReadJSON(&msg); err != nil {
			log.Println("websocket read error:", err)
			return
		}

		switch msg.Type {
		case "driver.location.updated":
			if err := c.handleLocation(ctx, msg); err != nil {
				log.Println("location update error:", err)
			}

		case "heartbeat":
			_ = c.driverCache.RefreshHeartbeat(ctx, c.driverID)
			_ = c.driverCache.RefreshConnection(ctx, c.driverID)

		case "ride.offer.ack":
			if err := c.handleOfferAck(ctx, msg); err != nil {
				log.Println("offer ack error:", err)
			}

		default:
			log.Println("unknown websocket message type:", msg.Type)
		}
	}
}

func (c *Client) writePump(ctx context.Context) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			if !ok {
				return
			}

			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))

			if err := c.conn.WriteJSON(msg); err != nil {
				return
			}

		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))

			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}

		case <-ctx.Done():
			return

		case <-c.done:
			return
		}
	}
}

func (c *Client) handleLocation(ctx context.Context, msg IncomingMessage) error {
	tracer := otel.Tracer("realtime.websocket")

	ctx, span := tracer.Start(ctx, "driver.location.updated")
	defer span.End()

	span.SetAttributes(
		attribute.String("driver.id", c.driverID.String()),
		attribute.Float64("location.lat", msg.Lat),
		attribute.Float64("location.lng", msg.Lng),
		attribute.Float64("location.accuracy", msg.Accuracy),
		attribute.Float64("location.speed", msg.Speed),
		attribute.Float64("location.bearing", msg.Bearing),
		attribute.Int64("location.seq", msg.Seq),
	)

	// Rate limiting
	if !c.lastLocationAt.IsZero() &&
		time.Since(c.lastLocationAt) < c.minLocationGap {

		span.SetAttributes(
			attribute.String("location.drop_reason", "rate_limited"),
		)

		span.SetStatus(codes.Ok, "dropped")
		return nil
	}

	err := c.driverLocationService.ProcessRealtimeLocation(
		ctx,
		c.driverID,
		msg.Lat,
		msg.Lng,
		msg.Accuracy,
		msg.Speed,
		msg.Bearing,
		msg.Seq,
	)

	switch {
	case errors.Is(err, application.ErrInvalidCoordinates):
		span.SetAttributes(
			attribute.String("location.drop_reason", "invalid_coordinates"),
		)
		span.SetStatus(codes.Ok, "dropped")
		return nil

	case errors.Is(err, application.ErrBadAccuracy):
		span.SetAttributes(
			attribute.String("location.drop_reason", "bad_accuracy"),
		)
		span.SetStatus(codes.Ok, "dropped")
		return nil

	case errors.Is(err, application.ErrInvalidSequence):
		span.SetAttributes(
			attribute.String("location.drop_reason", "invalid_sequence"),
		)
		span.SetStatus(codes.Ok, "dropped")
		return nil

	case errors.Is(err, application.ErrStaleSequence):
		span.SetAttributes(
			attribute.String("location.drop_reason", "stale_sequence"),
		)
		span.SetStatus(codes.Ok, "dropped")
		return nil

	case err != nil:
		// Unexpected infrastructure/system failure.
		// The child span already records the specific reason.
		span.RecordError(err)
		span.SetStatus(codes.Error, "processing_failed")
		return err
	}

	c.lastLocationAt = time.Now()

	span.SetAttributes(
		attribute.Bool("location.accepted", true),
	)

	span.SetStatus(codes.Ok, "location_updated")

	return nil
}

func marshalMessage(msg OutgoingMessage) []byte {
	b, _ := json.Marshal(msg)
	return b
}

func (c *Client) handleOfferAck(ctx context.Context, msg IncomingMessage) error {
	rideID, err := uuid.Parse(msg.RideID)
	if err != nil {
		return nil
	}

	driverID, err := uuid.Parse(msg.DriverID)
	if err != nil {
		return nil
	}

	if driverID != c.driverID {
		return nil
	}

	if err := c.driverCache.MarkOfferAcked(ctx, rideID, c.driverID); err != nil {
		return err
	}

	return c.driverCache.MarkOfferDeliveryStatus(
		ctx,
		rideID,
		c.driverID,
		ride.OfferDeliveryAcked,
	)
}
