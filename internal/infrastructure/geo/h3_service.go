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

func (s *H3Service) NeighborCells(lat, lng float64) ([]string, error) {
	if lat < -90 || lat > 90 || lng < -180 || lng > 180 {
		return nil, fmt.Errorf("invalid coordinates")
	}

	center, err := h3.LatLngToCell(
		h3.NewLatLng(lat, lng),
		s.resolution,
	)
	if err != nil {
		return nil, err
	}

	cells, err := h3.GridDisk(center, s.searchRing)
	if err != nil {
		return nil, err
	}

	result := make([]string, 0, len(cells))
	for _, c := range cells {
		result = append(result, c.String())
	}

	return result, nil
}
