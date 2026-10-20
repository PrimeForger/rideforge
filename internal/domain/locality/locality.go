package locality

import (
	"errors"
	"strings"

	"github.com/google/uuid"
)

type Status string

const (
	StatusDraft    Status = "DRAFT"
	StatusActive   Status = "ACTIVE"
	StatusInactive Status = "INACTIVE"
)

var (
	ErrInvalidLocality = errors.New("invalid locality")
	ErrInvalidStatus   = errors.New("invalid locality status")
)

type Locality struct {
	ID                     uuid.UUID
	AdministrativeRegionID uuid.UUID
	Name                   string
	Slug                   string
	BoundaryWKT            string
	Status                 Status
}

func New(id, administrativeRegionID uuid.UUID, name, slug, boundaryWKT string) (Locality, error) {
	l := Locality{ID: id, AdministrativeRegionID: administrativeRegionID, Name: name, Slug: slug, BoundaryWKT: boundaryWKT, Status: StatusDraft}
	if err := l.Validate(); err != nil {
		return Locality{}, err
	}
	return l, nil
}

func (l Locality) Validate() error {
	if l.ID == uuid.Nil || l.AdministrativeRegionID == uuid.Nil || strings.TrimSpace(l.Name) == "" || strings.TrimSpace(l.Slug) == "" || strings.TrimSpace(l.BoundaryWKT) == "" {
		return ErrInvalidLocality
	}
	switch l.Status {
	case StatusDraft, StatusActive, StatusInactive:
		return nil
	default:
		return ErrInvalidStatus
	}
}

func (l Locality) CanResolve() bool { return l.Status == StatusActive }
