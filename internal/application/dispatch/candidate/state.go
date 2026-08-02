package candidate

type State uint8

const (
	StateUnknown State = iota

	StateDiscovered

	StateFiltered

	StateReserved

	StateRanked

	StateOffered

	StateAccepted

	StateRejected
)
