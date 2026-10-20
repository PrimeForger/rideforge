package location

import (
	"math"
	"testing"
)

func TestCoordinate_IsValid(t *testing.T) {
	tests := []struct {
		name string
		lat  float64
		lng  float64
		want bool
	}{
		{name: "ordinary coordinate", lat: 12.9, lng: 77.5, want: true},
		{name: "minimum bounds", lat: -90, lng: -180, want: true},
		{name: "maximum bounds", lat: 90, lng: 180, want: true},
		{name: "equator and date line", lat: 0, lng: 180, want: true},
		{name: "null island", lat: 0, lng: 0, want: false},
		{name: "latitude below bounds", lat: -90.01, lng: 0, want: false},
		{name: "longitude above bounds", lat: 0, lng: 180.01, want: false},
		{name: "NaN latitude", lat: math.NaN(), lng: 1, want: false},
		{name: "infinite longitude", lat: 1, lng: math.Inf(1), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := (Coordinate{Lat: tt.lat, Lng: tt.lng}).IsValid(); got != tt.want {
				t.Errorf("Coordinate{Lat: %v, Lng: %v}.IsValid() = %v, want %v", tt.lat, tt.lng, got, tt.want)
			}
		})
	}
}
