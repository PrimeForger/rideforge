package profile

type ExpansionProfile struct {
	// Number of rings to advance after completing the current ring.
	RingStep int

	// Maximum ring that may be searched.
	MaxRing int

	// Maximum cells that may be visited.
	MaxCells int

	// Maximum number of candidate IDs that may be collected.
	MaxCandidates int
}
