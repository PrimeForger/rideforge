package locality

import (
	"github.com/google/uuid"
	"testing"
)

func TestLocalityValidation(t *testing.T) {
	id, regionID := uuid.New(), uuid.New()
	if _, err := New(id, regionID, "Kasaragod", "kasaragod", "MULTIPOLYGON EMPTY"); err != nil {
		t.Fatalf("valid locality rejected: %v", err)
	}
	if _, err := New(uuid.Nil, regionID, "name", "slug", "POINT EMPTY"); err == nil {
		t.Fatal("nil canonical ID accepted")
	}
}

func TestLocalityResolutionOnlyActive(t *testing.T) {
	l := Locality{ID: uuid.New(), AdministrativeRegionID: uuid.New(), Name: "x", Slug: "x", BoundaryWKT: "POINT EMPTY", Status: StatusDraft}
	if l.CanResolve() {
		t.Fatal("draft locality can resolve")
	}
	l.Status = StatusActive
	if !l.CanResolve() {
		t.Fatal("active locality cannot resolve")
	}
}
