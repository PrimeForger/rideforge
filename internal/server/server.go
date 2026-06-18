package server

import (
	"net/http"

	"github.com/ashadashraf/ride-hail-app/internal/application"
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
	driverDeviceService *application.DriverDeviceService,
	realtimeHub *realtime.Hub,
	geoService *redis.GeoService,
	driverCache *redis.DriverCache,
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
		),
	}
}

func (s *Server) RegisterRoutes() {

	// Websocket
	http.HandleFunc("ws/driver", s.webSocketHandler.DriverSocket)

	http.HandleFunc("/rides", s.rideHandler.CreateRide)

	http.HandleFunc("/driver/online", s.driverHandler.GoOnline)
	http.HandleFunc("/driver/offline", s.driverHandler.GoOffline)
	http.HandleFunc("/driver/location", s.driverHandler.UpdateLocation)

	http.HandleFunc("/driver/rides/accept", s.driverHandler.AcceptRide)
	http.HandleFunc("/driver/rides/reject", s.driverHandler.RejectRide)

	http.HandleFunc("/driver/push-token", s.driverHandler.RegisterPushToken)
}
