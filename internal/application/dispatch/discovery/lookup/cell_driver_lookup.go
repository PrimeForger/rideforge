package lookup

import (
	"context"

	"github.com/google/uuid"
)

type CellDriverLookup interface {
	GetDriversInCells(
		ctx context.Context,
		cells []string,
		limit int,
	) ([]uuid.UUID, error)
}
