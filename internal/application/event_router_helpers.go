package application

import (
	"encoding/json"
	"errors"

	appevents "github.com/ashadashraf/ride-hail-app/internal/application/events"
	"github.com/google/uuid"
)

type driverRideEventData struct {
	RideID   string `json:"ride_id"`
	DriverID string `json:"driver_id"`
}

func decodeEventData(envelope appevents.Envelope, dst any) error {
	raw, err := json.Marshal(envelope.Data)
	if err != nil {
		return err
	}

	return json.Unmarshal(raw, dst)
}

func parseUUID(raw string, field string) (uuid.UUID, error) {
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, errors.New("invalid " + field)
	}

	return id, nil
}

func parseRideDriverIDs(data driverRideEventData) (uuid.UUID, uuid.UUID, error) {
	rideID, err := parseUUID(data.RideID, "ride_id")
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}

	driverID, err := parseUUID(data.DriverID, "driver_id")
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}

	return rideID, driverID, nil
}
