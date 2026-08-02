package pipeline

import (
	"context"

	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/candidate"
)

type Pipeline interface {
	Execute(
		ctx context.Context,
		pipelineCtx *Context,
		candidates *candidate.Collection,
	) error
}
