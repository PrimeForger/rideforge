package events

import "github.com/google/uuid"

type MatchingStartedEvent struct {
	RideID uuid.UUID
}

func (e MatchingStartedEvent) Name() string {
	return "matching.started"
}

type MatchingRetryEvent struct {
	RideID uuid.UUID
}

func (e MatchingRetryEvent) Name() string {
	return "matching.retry"
}

type DriverOfferedEvent struct {
	RideID   uuid.UUID
	DriverID uuid.UUID
}

func (e DriverOfferedEvent) Name() string {
	return "driver.offered"
}

type DriverAcceptedEvent struct {
	RideID   uuid.UUID
	DriverID uuid.UUID
}

func (e DriverAcceptedEvent) Name() string {
	return "driver.accepted"
}

type DriverRejectedEvent struct {
	RideID   uuid.UUID
	DriverID uuid.UUID
}

func (e DriverRejectedEvent) Name() string {
	return "driver.rejected"
}

type DriverRejectedProcessedEvent struct {
	RideID   uuid.UUID
	DriverID uuid.UUID
}

func (e DriverRejectedProcessedEvent) Name() string {
	return "driver.rejected.processed"
}

type DriverTimeoutEvent struct {
	RideID   uuid.UUID `json:"ride_id"`
	DriverID uuid.UUID `json:"driver_id"`
}

func (e DriverTimeoutEvent) Name() string {
	return "driver.timeout"
}
