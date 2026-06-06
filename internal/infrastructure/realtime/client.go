package realtime

import (
	"context"
	"log"
	"time"

	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/redis"
	"github.com/google/uuid"
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

	send chan OutgoingMessage
	done chan struct{}
}

func NewClient(
	driverID uuid.UUID,
	conn *websocket.Conn,
	hub *Hub,
	geo *redis.GeoService,
	driverCache *redis.DriverCache,
) *Client {
	return &Client{
		driverID:    driverID,
		conn:        conn,
		hub:         hub,
		geo:         geo,
		driverCache: driverCache,
		send:        make(chan OutgoingMessage, sendBufferSize),
		done:        make(chan struct{}),
	}
}

func (c *Client) Start(ctx context.Context) {
	c.hub.Register(c.driverID, c)

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
		c.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		var msg IncomingMessage
		if err:=c.conn.ReadJSON(&msg)
	}
}
