package application

import (
	"math"
	"testing"
)

func TestCreateRideRequestValidatePickup(t *testing.T) {
	zero := float64(0)
	validLatitude := 91.0
	validLongitude := 181.0
	nan := math.NaN()
	infinity := math.Inf(1)
	tests := []struct {
		name    string
		lat     *float64
		lng     *float64
		wantErr bool
	}{
		{name: "missing latitude", lng: &zero, wantErr: true},
		{name: "missing longitude", lat: &zero, wantErr: true},
		{name: "valid zero coordinates", lat: &zero, lng: &zero},
		{name: "latitude out of range", lat: &validLatitude, lng: &zero, wantErr: true},
		{name: "longitude out of range", lat: &zero, lng: &validLongitude, wantErr: true},
		{name: "NaN latitude", lat: &nan, lng: &zero, wantErr: true},
		{name: "infinite longitude", lat: &zero, lng: &infinity, wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := (CreateRideRequest{PickupLatitude: test.lat, PickupLongitude: test.lng}).ValidatePickup()
			if (err != nil) != test.wantErr {
				t.Fatalf("ValidatePickup() error = %v, wantErr %v", err, test.wantErr)
			}
		})
	}
}
