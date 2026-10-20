package location

import "math"

type Coordinate struct {
	Lat float64
	Lng float64
}

// IsValid reports whether the coordinate is within geographic bounds and is
// not the (0, 0) GPS null-island sentinel.
func (c Coordinate) IsValid() bool {
	if math.IsNaN(c.Lat) || math.IsInf(c.Lat, 0) || math.IsNaN(c.Lng) || math.IsInf(c.Lng, 0) {
		return false
	}
	if c.Lat < -90 || c.Lat > 90 || c.Lng < -180 || c.Lng > 180 {
		return false
	}
	return c.Lat != 0 || c.Lng != 0
}

func (c Coordinate) DistanceTo(other Coordinate) float64 {
	// stub - replace with haversine later
	return 0
}
