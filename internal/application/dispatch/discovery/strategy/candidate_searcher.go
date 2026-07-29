package strategy

import (
	"context"

	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/discovery/search"
)

type CandidateSearcher interface {
	FindCandidates(ctx context.Context, req search.Request) (search.Result, error)
}
