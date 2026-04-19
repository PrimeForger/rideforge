package redis

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	rdb *redis.Client
}

func NewClient(addr string) (*Client, error) {

	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	// health check
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	return &Client{rdb: rdb}, nil
}

func (c *Client) GetRaw() *redis.Client {
	return c.rdb
}
