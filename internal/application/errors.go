package application

import "errors"

var (
	ErrInvalidCoordinates = errors.New("invalid coordinates")
	ErrBadAccuracy        = errors.New("bad accuracy")
	ErrInvalidSequence    = errors.New("invalid sequence")
	ErrStaleSequence      = errors.New("stale sequence")
)
