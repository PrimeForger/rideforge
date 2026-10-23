package location

import (
	"context"
	"math"
	"testing"

	"github.com/google/uuid"
)

func TestDriverLocationService_RejectsInvalidCoordinates(t *testing.T) {
	service := &DriverLocationService{}
	tests := []struct {
		name string
		lat  float64
		lng  float64
	}{
		{name: "null island", lat: 0, lng: 0},
		{name: "NaN latitude", lat: math.NaN(), lng: 1},
		{name: "infinite longitude", lat: 1, lng: math.Inf(1)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := service.ProcessRealtimeLocation(context.Background(), uuid.Nil, tt.lat, tt.lng, 1, 0, 0, 1); err != ErrInvalidCoordinates {
				t.Errorf("ProcessRealtimeLocation() error = %v, want %v", err, ErrInvalidCoordinates)
			}
			if err := service.PersistLocation(context.Background(), uuid.Nil, tt.lat, tt.lng); err != ErrInvalidCoordinates {
				t.Errorf("PersistLocation() error = %v, want %v", err, ErrInvalidCoordinates)
			}
		})
	}
}
