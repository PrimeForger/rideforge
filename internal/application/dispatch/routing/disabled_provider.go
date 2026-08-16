package routing

import "context"

type DisabledProvider struct{}

func NewDisabledProvider() *DisabledProvider {
	return &DisabledProvider{}
}

func (p *DisabledProvider) CalculateRoute(ctx context.Context, req RouteRequest) (RouteResult, error) {
	return RouteResult{}, ErrProviderDisabled
}
