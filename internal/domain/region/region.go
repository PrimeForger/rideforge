package region

type Region string

const (
	Kerala    Region = "KERALA"
	Karnataka Region = "KARNATAKA"
)

type BorderRule struct {
	From Region
	To   Region
}

func IsRideAllowed(from, to Region) bool {
	// critical legal rule layer
	if from == Kerala && to == Karnataka {
		return false
	}

	if from == Karnataka && to == Kerala {
		return false
	}

	return true
}
