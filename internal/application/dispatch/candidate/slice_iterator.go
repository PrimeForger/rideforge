package candidate

type SliceIterator struct {
	collection *Collection
	index      int
}

func NewSliceIterator(
	collection *Collection,
) *SliceIterator {

	return &SliceIterator{
		collection: collection,
	}
}

func (it *SliceIterator) Next() (*Candidate, bool) {

	if it.index >= len(it.collection.items) {
		return nil, false
	}

	candidate := &it.collection.items[it.index]

	it.index++

	return candidate, true
}

var _ Iterator = (*SliceIterator)(nil)
