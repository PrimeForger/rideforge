package ride

import "time"

const (
	EventRideRequested = "ride.requested"
	EventDriverMatched = "ride.driver_matched"
)

type RideRequested struct {
	RideID      string    `json:"ride_id"`
	Passenger   string    `json:"passenger"`
	Pickup      string    `json:"pickup"`
	Dropoff     string    `json:"dropoff"`
	RequestedAt time.Time `json:"requested_at"`
}

type DriverMatched struct {
	RideID    string    `json:"ride_id"`
	DriverID  string    `json:"driver_id"`
	MatchedAt time.Time `json:"matched_at"`
}
