package events

import "github.com/google/uuid"

type DriverPushTokenUpdatedEvent struct {
	DriverID uuid.UUID `json:"driver_id"`
	DeviceID string    `json:"device_id"`
	Platform string    `json:"platform"`
	Token    string    `json:"token"`
}

func (e DriverPushTokenUpdatedEvent) Name() string {
	return "driver.push_token.updated"
}
