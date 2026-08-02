package ranking

import (
	"context"

	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/candidate"
)

type Scorer interface {
	Score(
		ctx context.Context,
		features Features,
	) (candidate.Score, error)
}
