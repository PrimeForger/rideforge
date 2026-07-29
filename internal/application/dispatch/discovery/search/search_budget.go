package search

type SearchBudget struct {
	CandidateLimit int

	RemainingCandidates int

	MaxRing        int
	RemainingRings int

	MaxCells       int
	RemainingCells int
}

func NewSearchBudget(
	candidateLimit int,
	maxRing int,
) SearchBudget {

	return SearchBudget{
		CandidateLimit: candidateLimit,

		RemainingCandidates: candidateLimit,

		MaxRing:        maxRing,
		RemainingRings: maxRing + 1,
	}
}

func (b *SearchBudget) ConfigureLimits(
	maxRing int,
	maxCells int,
	maxCandidates int,
) {

	if maxRing > 0 {
		b.MaxRing = maxRing
		b.RemainingRings = maxRing + 1
	}

	if maxCells > 0 {
		b.MaxCells = maxCells
		b.RemainingCells = maxCells
	}

	if maxCandidates > 0 && maxCandidates < b.CandidateLimit {

		b.CandidateLimit = maxCandidates

		if b.RemainingCandidates > maxCandidates {
			b.RemainingCandidates = maxCandidates
		}
	}
}

func (b *SearchBudget) ConsumeCandidates(count int) {
	b.RemainingCandidates -= count

	if b.RemainingCandidates < 0 {
		b.RemainingCandidates = 0
	}
}

func (b *SearchBudget) ConsumeRing() {
	if b.RemainingRings > 0 {
		b.RemainingRings--
	}
}

func (b *SearchBudget) ConsumeCells(count int) {

	if b.MaxCells == 0 {
		return
	}

	b.RemainingCells -= count

	if b.RemainingCells < 0 {
		b.RemainingCells = 0
	}
}

func (b SearchBudget) CandidateBudgetExhausted() bool {
	return b.RemainingCandidates <= 0
}

func (b SearchBudget) RingBudgetExhausted() bool {
	return b.RemainingRings <= 0
}

func (b SearchBudget) CellBudgetExhausted() bool {
	return b.MaxCells > 0 &&
		b.RemainingCells <= 0
}
