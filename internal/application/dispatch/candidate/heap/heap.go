package heap

import (
	"container/heap"

	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/candidate"
)

type Heap struct {
	items []*candidate.Candidate
}

func NewHeap() *Heap {
	h := &Heap{}
	heap.Init(h)
	return h
}

func (h *Heap) Len() int {
	return len(h.items)
}

func (h *Heap) Less(i, j int) bool {
	// max heap
	if h.items[i].Metadata.Score == nil ||
		h.items[j].Metadata.Score == nil {

		panic("candidate score is nil")
	}

	return h.items[i].Metadata.Score.Value >
		h.items[j].Metadata.Score.Value
}

func (h *Heap) Swap(i, j int) {
	h.items[i], h.items[j] = h.items[j], h.items[i]
}

func (h *Heap) Push(x any) {

	c, ok := x.(*candidate.Candidate)
	if !ok {
		panic("heap.Push expects *candidate.Candidate")
	}

	h.items = append(h.items, c)
}

func (h *Heap) Pop() any {

	old := h.items

	n := len(old)

	item := old[n-1]

	h.items = old[:n-1]

	return item
}

func (h *Heap) PushCandidate(
	c *candidate.Candidate,
) {
	heap.Push(h, c)
}

func (h *Heap) PopCandidate() (*candidate.Candidate, bool) {

	if h.Len() == 0 {
		return nil, false
	}

	return heap.Pop(h).(*candidate.Candidate), true
}
