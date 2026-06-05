package server

import (
	"net/http"

	"github.com/ashadashraf/ride-hail-app/internal/application"
)

type Server struct {
	rideHandler   *RideHandler
	driverHandler *DriverHandler
}

func NewServer(
	rideService *application.RideService,
	driverService *application.DriverService,
	driverResponseCommandService *application.DriverResponseCommandService,
) *Server {

	return &Server{
		rideHandler: NewRideHandler(rideService),
		driverHandler: NewDriverHandler(
			driverService,
			driverResponseCommandService,
		),
	}
}

func (s *Server) RegisterRoutes() {
	http.HandleFunc("/rides", s.rideHandler.CreateRide)

	http.HandleFunc("/driver/online", s.driverHandler.GoOnline)
	http.HandleFunc("/driver/offline", s.driverHandler.GoOffline)
	http.HandleFunc("/driver/location", s.driverHandler.UpdateLocation)

	http.HandleFunc("/driver/rides/accept", s.driverHandler.AcceptRide)
	http.HandleFunc("/driver/rides/reject", s.driverHandler.RejectRide)
}
