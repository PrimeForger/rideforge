package selector

import (
	"context"

	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/discovery/pipeline"
	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/discovery/profile"
	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/discovery/search"
)

type DefaultSelector struct {
	pipeline pipeline.Pipeline
}

func NewDefaultSelector(
	pipeline pipeline.Pipeline,
) *DefaultSelector {

	return &DefaultSelector{
		pipeline: pipeline,
	}
}

func (s *DefaultSelector) Select(
	ctx context.Context,
	state *search.SearchState,
) (profile.SearchProfile, error) {

	builder := profile.NewBuilder()

	if err := s.pipeline.Execute(
		ctx,
		state,
		builder,
	); err != nil {

		return profile.SearchProfile{}, err
	}

	return builder.Build(), nil
}
