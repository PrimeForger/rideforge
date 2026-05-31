package server

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/ashadashraf/ride-hail-app/internal/application"
	"github.com/ashadashraf/ride-hail-app/internal/domain/region"
	"github.com/google/uuid"
)

type RideHandler struct {
	rideService *application.RideService
}

func NewRideHandler(
	rideService *application.RideService,
) *RideHandler {
	return &RideHandler{
		rideService: rideService,
	}
}

func (h *RideHandler) CreateRide(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		RiderID string `json:"rider_id"`
		From    string `json:"from_region"`
		To      string `json:"to_region"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	riderUUID, err := uuid.Parse(req.RiderID)
	if err != nil {
		http.Error(w, "invalid rider id", http.StatusBadRequest)
		return
	}

	rideID, err := h.rideService.CreateRide(
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

	writeJSON(w, http.StatusCreated, map[string]string{
		"ride_id": rideID.String(),
	})
}
