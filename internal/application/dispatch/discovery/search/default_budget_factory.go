package search

type DefaultBudgetFactory struct {
	maxRing int
}

func NewDefaultBudgetFactory(
	maxRing int,
) *DefaultBudgetFactory {

	return &DefaultBudgetFactory{
		maxRing: maxRing,
	}
}

func (f *DefaultBudgetFactory) NewBudget(
	input PolicyInput,
) SearchBudget {

	return NewSearchBudget(
		input.CandidateLimit,
		f.maxRing,
	)
}
