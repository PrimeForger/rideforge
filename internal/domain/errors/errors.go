package errors

import "errors"

var (
	ErrIllegalRideRegion = errors.New("illegal cross-region ride")
	ErrDriverUnavailable = errors.New("driver unavailable")
	ErrInvalidState      = errors.New("invalid state transition")
)
