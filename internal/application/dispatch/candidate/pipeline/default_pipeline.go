package pipeline

import (
	"context"

	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/candidate"
)

type DefaultPipeline struct {
	stages []Stage
}

func New(
	stages ...Stage,
) *DefaultPipeline {

	return &DefaultPipeline{
		stages: stages,
	}
}

func (p *DefaultPipeline) Execute(
	ctx context.Context,
	pipelineCtx *Context,
	candidates *candidate.Collection,
) error {

	for _, stage := range p.stages {

		if err := stage.Execute(
			ctx,
			pipelineCtx,
			candidates,
		); err != nil {
			return err
		}
	}

	return nil
}

var _ Pipeline = (*DefaultPipeline)(nil)
