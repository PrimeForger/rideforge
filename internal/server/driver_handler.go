package server

import (
	"net/http"

	"github.com/ashadashraf/ride-hail-app/internal/application"
	"github.com/google/uuid"
)

type DriverHandler struct {
	driverService   *application.DriverService
	responseService *application.DriverResponseCommandService
	deviceService   *application.DriverDeviceService
}

func NewDriverHandler(
	driverService *application.DriverService,
	responseService *application.DriverResponseCommandService,
	deviceService *application.DriverDeviceService,
) *DriverHandler {
	return &DriverHandler{
		driverService:   driverService,
		responseService: responseService,
		deviceService:   deviceService,
	}
}

func (h *DriverHandler) GoOnline(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	var req struct {
		DriverID string  `json:"driver_id"`
		Lat      float64 `json:"lat"`
		Lng      float64 `json:"lng"`
	}

	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	driverID, ok := parseUUID(w, req.DriverID, "driver_id")
	if !ok {
		return
	}

	if !validCoordinates(req.Lat, req.Lng) {
		writeError(w, http.StatusBadRequest, "invalid coordinates")
		return
	}

	if err := h.driverService.GoOnline(r.Context(), driverID, req.Lat, req.Lng); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusAccepted, map[string]string{
		"status":    "online",
		"driver_id": driverID.String(),
	})
}

func (h *DriverHandler) GoOffline(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	var req struct {
		DriverID string `json:"driver_id"`
	}

	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	driverID, ok := parseUUID(w, req.DriverID, "driver_id")
	if !ok {
		return
	}

	if err := h.driverService.GoOffline(r.Context(), driverID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusAccepted, map[string]string{
		"status":    "offline",
		"driver_id": driverID.String(),
	})
}

func (h *DriverHandler) UpdateLocation(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	var req struct {
		DriverID string  `json:"driver_id"`
		Lat      float64 `json:"lat"`
		Lng      float64 `json:"lng"`
	}

	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	driverID, ok := parseUUID(w, req.DriverID, "driver_id")
	if !ok {
		return
	}

	if !validCoordinates(req.Lat, req.Lng) {
		writeError(w, http.StatusBadRequest, "invalid coordinates")
		return
	}

	if err := h.driverService.UpdateLocation(r.Context(), driverID, req.Lat, req.Lng); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusAccepted, map[string]string{
		"status":    "location_updated",
		"driver_id": driverID.String(),
	})
}

func (h *DriverHandler) AcceptRide(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	rideID, driverID, ok := h.parseRideResponseRequest(w, r)
	if !ok {
		return
	}

	if err := h.responseService.AcceptRide(r.Context(), rideID, driverID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusAccepted, map[string]string{
		"status":    "accepted",
		"ride_id":   rideID.String(),
		"driver_id": driverID.String(),
	})
}

func (h *DriverHandler) RejectRide(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	rideID, driverID, ok := h.parseRideResponseRequest(w, r)
	if !ok {
		return
	}

	if err := h.responseService.RejectRide(r.Context(), rideID, driverID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusAccepted, map[string]string{
		"status":    "rejected",
		"ride_id":   rideID.String(),
		"driver_id": driverID.String(),
	})
}

func (h *DriverHandler) parseRideResponseRequest(
	w http.ResponseWriter,
	r *http.Request,
) (uuid.UUID, uuid.UUID, bool) {
	var req struct {
		RideID   string `json:"ride_id"`
		DriverID string `json:"driver_id"`
	}

	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return uuid.Nil, uuid.Nil, false
	}

	rideID, ok := parseUUID(w, req.RideID, "ride_id")
	if !ok {
		return uuid.Nil, uuid.Nil, false
	}

	driverID, ok := parseUUID(w, req.DriverID, "driver_id")
	if !ok {
		return uuid.Nil, uuid.Nil, false
	}

	return rideID, driverID, true
}

func (h *DriverHandler) RegisterPushToken(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	var req struct {
		DriverID string `json:"driver_id"`
		DeviceID string `json:"device_id"`
		Platform string `json:"platform"`
		Token    string `json:"token"`
	}

	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	driverID, ok := parseUUID(w, req.DriverID, "driver_id")
	if !ok {
		return
	}

	if err := h.deviceService.RegisterPushToken(
		r.Context(),
		driverID,
		req.DeviceID,
		req.Platform,
		req.Token,
	); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusAccepted, map[string]string{
		"status":    "push_token_registered",
		"driver_id": driverID.String(),
	})
}
