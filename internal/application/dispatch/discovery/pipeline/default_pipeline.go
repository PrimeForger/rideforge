package pipeline

import (
	"context"

	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/discovery/policy"
	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/discovery/profile"
	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/discovery/search"
)

type DefaultPipeline struct {
	stages []policy.SearchProfileRule
}

func New(
	stages ...policy.SearchProfileRule,
) *DefaultPipeline {

	return &DefaultPipeline{
		stages: stages,
	}
}

func (p *DefaultPipeline) Execute(
	ctx context.Context,
	state *search.SearchState,
	builder *profile.Builder,
) error {

	for _, stage := range p.stages {

		if err := stage.Apply(
			ctx,
			state,
			builder,
		); err != nil {

			return err
		}
	}

	return nil
}
