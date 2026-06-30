package server

import (
	"net/http"

	"github.com/ashadashraf/ride-hail-app/internal/application"
	"github.com/ashadashraf/ride-hail-app/internal/config"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/realtime"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/redis"
)

type Server struct {
	rideHandler      *RideHandler
	driverHandler    *DriverHandler
	webSocketHandler *WebSocketHandler
}

func NewServer(
	rideService *application.RideService,
	driverService *application.DriverService,
	driverResponseCommandService *application.DriverResponseCommandService,
	driverLocationService *application.DriverLocationService,
	driverDeviceService *application.DriverDeviceService,
	realtimeHub *realtime.Hub,
	geoService *redis.GeoService,
	driverCache *redis.DriverCache,
	realtimeCfg *config.RealtimeConfig,
) *Server {

	return &Server{
		rideHandler: NewRideHandler(rideService),
		driverHandler: NewDriverHandler(
			driverService,
			driverResponseCommandService,
			driverDeviceService,
		),
		webSocketHandler: NewWebSocketHandler(
			realtimeHub,
			geoService,
			driverCache,
			driverLocationService,
			realtimeCfg,
		),
	}
}

func (s *Server) RegisterRoutes(mux *http.ServeMux) {

	// Websocket
	mux.HandleFunc("ws/driver", s.webSocketHandler.DriverSocket)

	mux.HandleFunc("/rides", s.rideHandler.CreateRide)

	mux.HandleFunc("/driver/online", s.driverHandler.GoOnline)
	mux.HandleFunc("/driver/offline", s.driverHandler.GoOffline)
	mux.HandleFunc("/driver/location", s.driverHandler.UpdateLocation)

	mux.HandleFunc("/driver/rides/accept", s.driverHandler.AcceptRide)
	mux.HandleFunc("/driver/rides/reject", s.driverHandler.RejectRide)

	mux.HandleFunc("/driver/push-token", s.driverHandler.RegisterPushToken)
}
