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
	RideID          uuid.UUID `json:"ride_id"`
	DriverID        uuid.UUID `json:"driver_id"`
	OfferTimeoutMs  int64     `json:"offer_timeout_ms"`
	MatchingAttempt int       `json:"matching_attempt"`
	SearchRadiusKm  float64   `json:"search_radius_km"`
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

type DriverTimeoutProcessedEvent struct {
	RideID         uuid.UUID `json:"ride_id"`
	DriverID       uuid.UUID `json:"driver_id"`
	OfferAcked     bool      `json:"offer_acked"`
	DeliveryStatus string    `json:"delivery_status"`
	TimeoutReason  string    `json:"timeout_reason"`
}

func (e DriverTimeoutProcessedEvent) Name() string {
	return "driver.timeout.processed"
}

type DriverOfferAckedEvent struct {
	RideID   uuid.UUID `json:"ride_id"`
	DriverID uuid.UUID `json:"driver_id"`
	AckedAt  int64     `json:"acked_at"`
}

func (e DriverOfferAckedEvent) Name() string {
	return "driver.offer.acked"
}
