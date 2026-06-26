package server

import (
	"net/http"

	"github.com/ashadashraf/ride-hail-app/internal/config"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/realtime"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/redis"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type WebSocketHandler struct {
	hub         *realtime.Hub
	geo         *redis.GeoService
	driverCache *redis.DriverCache
	cfg         *config.RealtimeConfig
}

func NewWebSocketHandler(
	hub *realtime.Hub,
	geo *redis.GeoService,
	driverCache *redis.DriverCache,
	cfg *config.RealtimeConfig,
) *WebSocketHandler {
	return &WebSocketHandler{
		hub:         hub,
		geo:         geo,
		driverCache: driverCache,
		cfg:         cfg,
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,

	CheckOrigin: func(r *http.Request) bool {
		// TODO: replace with strict allowed origins
		return true
	},
}

func (h *WebSocketHandler) DriverSocket(w http.ResponseWriter, r *http.Request) {
	driverIDRaw := r.URL.Query().Get("driver_id")
	driverID, err := uuid.Parse(driverIDRaw)
	if err != nil {
		http.Error(w, "invalid driver_id", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	client := realtime.NewClient(
		driverID,
		conn,
		h.hub,
		h.geo,
		h.driverCache,
		h.cfg,
	)

	client.Start(r.Context())
}

// For now auth placeholder is query param. Later replace with JWT.
