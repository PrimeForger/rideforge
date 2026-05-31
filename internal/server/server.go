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
	driverResponseCommandService *application.DriverResponseCommandService,
) *Server {

	return &Server{
		rideHandler:   NewRideHandler(rideService),
		driverHandler: NewDriverHandler(driverResponseCommandService),
	}
}

func (s *Server) RegisterRoutes() {
	http.HandleFunc("/rides", s.rideHandler.CreateRide)
	http.HandleFunc("/driver/rides/accept", s.driverHandler.AcceptRide)
	http.HandleFunc("/driver/rides/reject", s.driverHandler.RejectRide)
}
