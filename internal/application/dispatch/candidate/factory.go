package candidate

import "github.com/google/uuid"

func New(
	id uuid.UUID,
) Candidate {

	return Candidate{
		ID: id,

		State: StateDiscovered,
	}
}

func NewCollectionFromIDs(
	ids []uuid.UUID,
) *Collection {

	collection := NewCollection(len(ids))

	for _, id := range ids {
		collection.Add(New(id))
	}

	return collection
}
