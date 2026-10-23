package httptransport

import "net/http"

func (s *Server) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("ws/driver", s.webSocketHandler.DriverSocket)
	mux.HandleFunc("/rides", s.rideHandler.CreateRide)
	mux.HandleFunc("/driver/online", s.driverHandler.GoOnline)
	mux.HandleFunc("/driver/offline", s.driverHandler.GoOffline)
	mux.HandleFunc("/driver/location", s.driverHandler.UpdateLocation)
	mux.HandleFunc("/driver/rides/accept", s.driverHandler.AcceptRide)
	mux.HandleFunc("/driver/rides/reject", s.driverHandler.RejectRide)
	mux.HandleFunc("/driver/push-token", s.driverHandler.RegisterPushToken)
}
