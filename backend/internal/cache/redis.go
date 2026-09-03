// Package cache wraps the shared Redis instance used for distributed rate
// limiting and refresh/access token blacklisting. Every key is namespaced
// with REDIS_PREFIX so this project never collides with other projects on
// the same shared Redis container.
package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"go-cp-boilerplate/internal/config"
)

type RedisClient struct {
	*redis.Client
	Prefix string
}

func NewRedisClient(cfg *config.Config) *RedisClient {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
	return &RedisClient{Client: client, Prefix: cfg.RedisPrefix}
}

// Key namespaces a raw key with the project's Redis prefix.
func (r *RedisClient) Key(parts ...string) string {
	key := r.Prefix
	for i, p := range parts {
		if i > 0 {
			key += ":"
		}
		key += p
	}
	return key
}

func (r *RedisClient) BlacklistToken(ctx context.Context, jti string, ttl time.Duration) error {
	return r.Set(ctx, r.Key("blacklist", jti), "1", ttl).Err()
}

func (r *RedisClient) IsBlacklisted(ctx context.Context, jti string) (bool, error) {
	n, err := r.Exists(ctx, r.Key("blacklist", jti)).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (r *RedisClient) Ping(ctx context.Context) error {
	return r.Client.Ping(ctx).Err()
}
