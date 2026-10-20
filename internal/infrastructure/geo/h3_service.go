package geo

import (
	"encoding/json"
	"fmt"
	"sort"

	h3 "github.com/uber/h3-go/v4"
)

type H3Service struct {
	resolution int
	searchRing int
}

// CellsForGeoJSON converts a PostGIS GeoJSON Polygon or MultiPolygon boundary
// into H3 cells at the service's configured operational resolution.
func (s *H3Service) CellsForGeoJSON(raw json.RawMessage) ([]string, error) {
	var geometry struct {
		Type        string          `json:"type"`
		Coordinates json.RawMessage `json:"coordinates"`
	}
	if err := json.Unmarshal(raw, &geometry); err != nil {
		return nil, fmt.Errorf("decode locality GeoJSON: %w", err)
	}
	var polygons [][][][]float64
	switch geometry.Type {
	case "Polygon":
		var polygon [][][]float64
		if err := json.Unmarshal(geometry.Coordinates, &polygon); err != nil {
			return nil, fmt.Errorf("decode locality polygon: %w", err)
		}
		polygons = [][][][]float64{polygon}
	case "MultiPolygon":
		if err := json.Unmarshal(geometry.Coordinates, &polygons); err != nil {
			return nil, fmt.Errorf("decode locality multipolygon: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported locality geometry type %q", geometry.Type)
	}

	set := make(map[string]struct{})
	for _, polygon := range polygons {
		if len(polygon) == 0 {
			return nil, fmt.Errorf("locality polygon has no exterior ring")
		}
		loops := make([]h3.GeoLoop, 0, len(polygon))
		for _, ring := range polygon {
			if len(ring) < 4 {
				return nil, fmt.Errorf("locality polygon ring has fewer than four coordinates")
			}
			loop := make(h3.GeoLoop, 0, len(ring))
			for _, coordinate := range ring {
				if len(coordinate) < 2 || coordinate[0] < -180 || coordinate[0] > 180 || coordinate[1] < -90 || coordinate[1] > 90 {
					return nil, fmt.Errorf("locality polygon contains invalid coordinate")
				}
				loop = append(loop, h3.NewLatLng(coordinate[1], coordinate[0]))
			}
			loops = append(loops, loop)
		}
		cells, err := h3.PolygonToCells(h3.GeoPolygon{GeoLoop: loops[0], Holes: loops[1:]}, s.resolution)
		if err != nil {
			return nil, fmt.Errorf("polyfill locality boundary: %w", err)
		}
		for _, cell := range cells {
			set[cell.String()] = struct{}{}
		}
	}
	if len(set) == 0 {
		return nil, fmt.Errorf("locality boundary produced no H3 cells")
	}
	result := make([]string, 0, len(set))
	for cell := range set {
		result = append(result, cell)
	}
	sort.Strings(result)
	return result, nil
}

func NewH3Service(
	resolution int,
	searchRing int,
) *H3Service {
	return &H3Service{
		resolution: resolution,
		searchRing: searchRing,
	}
}

// ----------------------------------------------------------------------
// Resolution / configuration
// ----------------------------------------------------------------------

func (s *H3Service) Resolution() int {
	return s.resolution
}

func (s *H3Service) MaxSearchRing() int {
	return s.searchRing
}

func (s *H3Service) CenterCell(
	lat, lng float64,
) (string, error) {
	return s.CellForLocation(lat, lng)
}

func (s *H3Service) CellForLocation(lat, lng float64) (string, error) {
	if lat < -90 || lat > 90 || lng < -180 || lng > 180 {
		return "", fmt.Errorf("invalid coordinates")
	}

	cell, err := h3.LatLngToCell(
		h3.NewLatLng(lat, lng),
		s.resolution,
	)
	if err != nil {
		return "", err
	}

	return cell.String(), nil
}

func (s *H3Service) CellFromString(cell string) h3.Cell {
	return h3.CellFromString(cell)
}

// CellsInRing returns only the cells that belong to the requested ring.
//
// ring = 0
//
//	center only
//
// ring = 1
//
//	first ring only
//
// ring = 2
//
//	second ring only
func (s *H3Service) CellsInRing(center string, ring int) ([]string, error) {

	if ring < 0 {
		return nil, fmt.Errorf("ring cannot be negative")
	}

	centerCell := s.CellFromString(center)
	if !centerCell.IsValid() {
		return nil, fmt.Errorf("invalid H3 cell: %s", centerCell)
	}

	if ring == 0 {
		return []string{center}, nil
	}

	cells, err := h3.GridRing(centerCell, ring)
	if err != nil {
		return nil, err
	}

	result := make([]string, len(cells))

	for i, cell := range cells {
		result[i] = cell.String()
	}

	return result, nil
}

// DiskCells returns every cell from the center until the requested ring.
//
// ring = 2
//
// returns
//
// center
// ring1
// ring2
func (s *H3Service) DiskCells(
	center string,
	ring int,
) ([]string, error) {

	if ring < 0 {
		return nil, fmt.Errorf("ring cannot be negative")
	}

	centerCell := s.CellFromString(center)
	if !centerCell.IsValid() {
		return nil, fmt.Errorf("invalid H3 cell: %s", centerCell)
	}

	cells, err := h3.GridDisk(centerCell, ring)
	if err != nil {
		return nil, err
	}

	result := make([]string, len(cells))

	for i, cell := range cells {
		result[i] = cell.String()
	}

	return result, nil
}
