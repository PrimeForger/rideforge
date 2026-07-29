package contract

type SearchTerminationReason int

const (
	SearchContinue SearchTerminationReason = iota

	SearchCandidateBudgetExhausted
	SearchRingBudgetExhausted
	SearchCellBudgetExhausted
)

func (r SearchTerminationReason) String() string {

	switch r {

	case SearchContinue:
		return "continue"

	case SearchCandidateBudgetExhausted:
		return "candidate_budget_exhausted"

	case SearchRingBudgetExhausted:
		return "ring_budget_exhausted"

	case SearchCellBudgetExhausted:
		return "cell_budget_exhausted"

	default:
		return "unknown"
	}
}
