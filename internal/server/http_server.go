package server

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/ashadashraf/ride-hail-app/internal/application"
	"github.com/ashadashraf/ride-hail-app/internal/domain/region"
	"github.com/google/uuid"
)

type Server struct {
	rideService *application.RideService
}

func NewServer(rideService *application.RideService) *Server {
	return &Server{
		rideService: rideService,
	}
}

func (s *Server) RegisterRoutes() {
	http.HandleFunc("/rides", s.createRide)
}

func (s *Server) createRide(w http.ResponseWriter, r *http.Request) {

	var req struct {
		RiderID string `json:"rider_id"`
		From    string `json:"from_region"`
		To      string `json:"to_region"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	riderUUID, err := uuid.Parse(req.RiderID)
	if err != nil {
		http.Error(w, "invalid rider id", http.StatusBadRequest)
		return
	}

	rideID, err := s.rideService.CreateRide(
		context.Background(),
		application.CreateRideRequest{
			RiderID: riderUUID,
		},
		region.Region(req.From),
		region.Region(req.To),
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := map[string]string{
		"ride_id": rideID.String(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
