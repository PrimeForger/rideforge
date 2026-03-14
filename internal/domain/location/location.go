package location

type Coordinate struct {
	Lat float64
	Lng float64
}

func (c Coordinate) DistanceTo(other Coordinate) float64 {
	// stub - replace with haversine later
	return 0
}
