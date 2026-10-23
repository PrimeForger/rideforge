package httptransport

import (
	"github.com/ashadashraf/ride-hail-app/internal/application"
	applicationdriver "github.com/ashadashraf/ride-hail-app/internal/application/driver"
	applicationlocation "github.com/ashadashraf/ride-hail-app/internal/application/location"
	"github.com/ashadashraf/ride-hail-app/internal/config"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/realtime"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/redis"
	"github.com/ashadashraf/ride-hail-app/internal/transport/http/handlers"
)

type Server struct {
	rideHandler      *handlers.RideHandler
	driverHandler    *handlers.DriverHandler
	webSocketHandler *handlers.WebSocketHandler
}

func NewServer(
	rideService *application.RideService,
	driverService *applicationdriver.DriverService,
	driverResponseCommandService *applicationdriver.DriverResponseCommandService,
	driverLocationService *applicationlocation.DriverLocationService,
	driverDeviceService *applicationdriver.DriverDeviceService,
	realtimeHub *realtime.Hub,
	geoService *redis.GeoService,
	driverCache *redis.DriverCache,
	realtimeCfg *config.RealtimeConfig,
) *Server {

	return &Server{
		rideHandler: handlers.NewRideHandler(rideService),
		driverHandler: handlers.NewDriverHandler(
			driverService,
			driverResponseCommandService,
			driverDeviceService,
		),
		webSocketHandler: handlers.NewWebSocketHandler(
			realtimeHub,
			geoService,
			driverCache,
			driverLocationService,
			realtimeCfg,
		),
	}
}
