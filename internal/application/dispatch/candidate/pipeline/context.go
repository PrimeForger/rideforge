package pipeline

import (
	"time"

	"github.com/google/uuid"
)

type Context struct {
	RideID uuid.UUID

	PickupLat float64
	PickupLng float64

	RetryAttempt int

	SearchRadiusKm float64
	CandidateLimit int

	StartedAt time.Time

	Result Result
}

func (c *Context) Duration() time.Duration {
	return time.Since(c.StartedAt)
}
