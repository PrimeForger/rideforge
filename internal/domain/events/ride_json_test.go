package events

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

func TestRideEventJSONContract(t *testing.T) {
	rideID := uuid.New()
	driverID := uuid.New()
	tests := []struct {
		name  string
		value any
		want  string
	}{
		{
			name:  "ride requested",
			value: RideRequestedEvent{RideID: rideID, PickupLatitude: 0, PickupLongitude: 0},
			want:  `{"ride_id":"` + rideID.String() + `","pickup_latitude":0,"pickup_longitude":0}`,
		},
		{
			name:  "ride accepted",
			value: RideAcceptedEvent{RideID: rideID, DriverID: driverID},
			want:  `{"ride_id":"` + rideID.String() + `","driver_id":"` + driverID.String() + `"}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := json.Marshal(test.value)
			if err != nil {
				t.Fatalf("json.Marshal() error = %v", err)
			}
			if string(got) != test.want {
				t.Fatalf("json.Marshal() = %s, want %s", got, test.want)
			}
		})
	}
}
