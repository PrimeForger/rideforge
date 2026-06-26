package driver

import "github.com/google/uuid"

type DriverMetricsSnapshot struct {
	DriverID         uuid.UUID
	AcceptanceRate   float64
	CancellationRate float64
	TimeoutRate      float64
	CompletedRides   int64
}
