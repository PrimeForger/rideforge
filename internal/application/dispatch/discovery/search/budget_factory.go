package search

type BudgetFactory interface {
	NewBudget(input PolicyInput) SearchBudget
}
