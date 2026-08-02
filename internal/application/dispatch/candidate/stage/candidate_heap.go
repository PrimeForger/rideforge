package stage

import "github.com/ashadashraf/ride-hail-app/internal/application/dispatch/candidate"

type CandidateHeap interface {
	Push(
		candidate candidate.Candidate,
		score float64,
	)
}
