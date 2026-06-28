package events

import "github.com/google/uuid"

// Driver Online
type DriverOnlineEvent struct {
	DriverID uuid.UUID `json:"driver_id"`
	Lat      float64   `json:"lat"`
	Lng      float64   `json:"lng"`
}

func (e DriverOnlineEvent) Name() string {
	return "driver.online"
}

// Driver Offline
type DriverOfflineEvent struct {
	DriverID uuid.UUID `json:"driver_id"`
	Reason   string    `json:"reason"`
}

func (e DriverOfflineEvent) Name() string {
	return "driver.offline"
}
