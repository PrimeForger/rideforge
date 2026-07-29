package redis

import (
	"context"

	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/geo"
	goredis "github.com/redis/go-redis/v9"
)

type H3DensityProvider struct {
	client *Client
	h3     *geo.H3Service
}

func NewH3DensityProvider(
	client *Client,
	h3 *geo.H3Service,
) *H3DensityProvider {

	return &H3DensityProvider{
		client: client,
		h3:     h3,
	}
}

func (p *H3DensityProvider) DriverCountInRing(
	ctx context.Context,
	centerCell string,
	ring int,
) (int, error) {

	cells, err := p.h3.CellsInRing(centerCell, ring)
	if err != nil {
		return 0, err
	}

	pipe := p.client.GetRaw().Pipeline()

	cmds := make([]*goredis.IntCmd, 0, len(cells))

	for _, cell := range cells {
		cmds = append(
			cmds,
			pipe.SCard(
				ctx,
				h3CellDriversKey(cell),
			),
		)
	}

	if _, err := pipe.Exec(ctx); err != nil {
		return 0, err
	}

	total := 0

	for _, cmd := range cmds {

		n, err := cmd.Result()
		if err != nil {
			return 0, err
		}

		total += int(n)
	}

	return total, nil
}
