package ride

type Status string

const (
	StatusRequested Status = "REQUESTED"
	StatusMatching  Status = "MATCHING"
	StatusAccepted  Status = "ACCEPTED"
	StatusArrived   Status = "ARRIVED"
	StatusStarted   Status = "STARTED"
	StatusCompleted Status = "COMPLETED"
	StatusCancelled Status = "CANCELLED"
)

func (s Status) CanTransitionTo(next Status) bool {
	switch s {
	case StatusRequested:
		return next == StatusMatching || next == StatusCancelled
	case StatusMatching:
		return next == StatusAccepted || next == StatusCancelled
	case StatusAccepted:
		return next == StatusArrived || next == StatusCancelled
	case StatusArrived:
		return next == StatusStarted || next == StatusCancelled
	case StatusStarted:
		return next == StatusCompleted || next == StatusCancelled
	default:
		return false
	}
}
