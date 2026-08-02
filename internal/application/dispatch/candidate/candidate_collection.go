package candidate

import "github.com/google/uuid"

type Collection struct {
	items []Candidate
}

func NewCollection(
	capacity int,
) *Collection {

	return &Collection{
		items: make([]Candidate, 0, capacity),
	}
}

func (c *Collection) Add(
	candidate Candidate,
) {
	c.items = append(c.items, candidate)
}

func (c *Collection) Len() int {
	return len(c.items)
}

func (c *Collection) Iterator() Iterator {
	return NewSliceIterator(c)
}

func (c *Collection) ForEach(
	fn func(*Candidate),
) {
	for i := range c.items {
		fn(&c.items[i])
	}
}

func (c *Collection) FindByID(
	id uuid.UUID,
) *Candidate {

	for i := range c.items {

		if c.items[i].ID == id {
			return &c.items[i]
		}
	}

	return nil
}

func (c *Collection) RemoveIf(
	predicate func(*Candidate) bool,
) {

	filtered := c.items[:0]

	for i := range c.items {

		if predicate(&c.items[i]) {
			continue
		}

		filtered = append(filtered, c.items[i])
	}

	c.items = filtered
}
