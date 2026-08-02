package pipeline

import "github.com/ashadashraf/ride-hail-app/internal/application/dispatch/candidate"

type CandidateSource interface {
	Next() (*candidate.Candidate, bool)
}
