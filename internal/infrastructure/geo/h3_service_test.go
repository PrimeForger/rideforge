package geo

import (
	"encoding/json"
	"testing"
)

func TestCellsForGeoJSONUsesConfiguredResolutionAndIsDeterministic(t *testing.T) {
	service := NewH3Service(8, 1)
	geometry := json.RawMessage(`{"type":"Polygon","coordinates":[[[74.98,12.48],[75.02,12.48],[75.02,12.52],[74.98,12.52],[74.98,12.48]]]}`)
	first, err := service.CellsForGeoJSON(geometry)
	if err != nil {
		t.Fatal(err)
	}
	second, err := service.CellsForGeoJSON(geometry)
	if err != nil {
		t.Fatal(err)
	}
	if len(first) == 0 || len(first) != len(second) {
		t.Fatalf("cell count mismatch: %d, %d", len(first), len(second))
	}
	for i := range first {
		if first[i] != second[i] {
			t.Fatalf("cell order differs at %d: %q != %q", i, first[i], second[i])
		}
		cell := service.CellFromString(first[i])
		if !cell.IsValid() || cell.Resolution() != 8 {
			t.Fatalf("cell %q is invalid or at wrong resolution", first[i])
		}
	}
}

func TestCellsForGeoJSONRejectsInvalidGeometry(t *testing.T) {
	service := NewH3Service(8, 1)
	for _, geometry := range []json.RawMessage{
		json.RawMessage(`{"type":"Point","coordinates":[0,0]}`),
		json.RawMessage(`{"type":"Polygon","coordinates":[[[181,0],[181,1],[180,1],[181,0]]]}`),
	} {
		if _, err := service.CellsForGeoJSON(geometry); err == nil {
			t.Fatalf("accepted invalid geometry %s", geometry)
		}
	}
}
