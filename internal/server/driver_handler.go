package server

import (
	"encoding/json"
	"net/http"

	"github.com/ashadashraf/ride-hail-app/internal/application"
	"github.com/google/uuid"
)

type DriverHandler struct {
	service *application.DriverResponseCommandService
}

func NewDriverHandler(
	service *application.DriverResponseCommandService,
) *DriverHandler {
	return &DriverHandler{service: service}
}

func (h *DriverHandler) AcceptRide(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RideID   string `json:"ride_id"`
		DriverID string `json:"driver_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	rideID, err := uuid.Parse(req.RideID)
	if err != nil {
		http.Error(w, "invalid ride_id", http.StatusBadRequest)
		return
	}

	driverID, err := uuid.Parse(req.DriverID)
	if err != nil {
		http.Error(w, "invalid driver_id", http.StatusBadRequest)
		return
	}

	if err := h.service.AcceptRide(r.Context(), rideID, driverID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusAccepted, map[string]string{
		"status":  "accepted",
		"ride_id": rideID.String(),
	})
}

func (h *DriverHandler) RejectRide(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RideID   string `json:"ride_id"`
		DriverID string `json:"driver_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	rideID, err := uuid.Parse(req.RideID)
	if err != nil {
		http.Error(w, "invalid ride_id", http.StatusBadRequest)
		return
	}

	driverID, err := uuid.Parse(req.DriverID)
	if err != nil {
		http.Error(w, "invalid driver_id", http.StatusBadRequest)
		return
	}

	if err := h.service.RejectRide(r.Context(), rideID, driverID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusAccepted, map[string]string{
		"status":  "rejected",
		"ride_id": rideID.String(),
	})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
