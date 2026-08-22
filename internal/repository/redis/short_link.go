package redis

import (
	"context"
	"errors"
	"time"

	redisclient "github.com/redis/go-redis/v9"
	"github.com/willfreit4s/short_link/internal/ports"
)

const shortLinkCacheKeyPrefix = "shortlink:"

type shortLinkCache struct {
	client *redisclient.Client
	ttl    time.Duration
}

func NewShortLinkCache(client *redisclient.Client, ttl time.Duration) ports.ShortLinkCache {
	return &shortLinkCache{client: client, ttl: ttl}
}

func (c *shortLinkCache) Get(ctx context.Context, hash string) (string, bool, error) {
	value, err := c.client.Get(ctx, c.key(hash)).Result()
	if err != nil {
		if errors.Is(err, redisclient.Nil) {
			return "", false, nil
		}

		return "", false, err
	}

	return value, true, nil
}

func (c *shortLinkCache) Set(ctx context.Context, hash string, originalURL string) error {
	ttl := c.ttl
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}

	return c.client.Set(ctx, c.key(hash), originalURL, ttl).Err()
}

func (c *shortLinkCache) Delete(ctx context.Context, hash string) error {
	return c.client.Del(ctx, c.key(hash)).Err()
}

func (c *shortLinkCache) key(hash string) string {
	return shortLinkCacheKeyPrefix + hash
}