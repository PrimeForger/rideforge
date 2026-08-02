package pipeline

type Result struct {
	Candidates CandidateSource

	LoadedCandidates   int
	FilteredCandidates int
	RankedCandidates   int
}
