package search

type PolicyInput struct {

	// Matching attempt (0 for first attempt).
	MatchingAttempt int

	// Maximum candidates requested by caller.
	CandidateLimit int
}
