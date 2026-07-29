package geo

import (
	"fmt"

	h3 "github.com/uber/h3-go/v4"
)

type H3Service struct {
	resolution int
	searchRing int
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
