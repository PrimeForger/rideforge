package candidate

import (
	"github.com/ashadashraf/ride-hail-app/internal/domain/driver"

	"github.com/google/uuid"
)

type Candidate struct {
	ID uuid.UUID

	State State

	Driver *driver.Driver

	Metadata Metadata
}
