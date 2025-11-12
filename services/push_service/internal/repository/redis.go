package repository

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisRepository offers small helpers around Redis for caching invalid tokens or rate limits.
type RedisRepository struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisRepository(client *redis.Client, ttl time.Duration) *RedisRepository {
	return &RedisRepository{
		client: client,
		ttl:    ttl,
	}
}

func (r *RedisRepository) Close() error {
	return r.client.Close()
}

// IsTokenSuppressed returns true if the token is currently marked as invalid.
func (r *RedisRepository) IsTokenSuppressed(ctx context.Context, token string) (bool, error) {
	key := "push:token:suppressed:" + token
	exists, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return exists == 1, nil
}

// SuppressToken stores a token in Redis with a TTL.
func (r *RedisRepository) SuppressToken(ctx context.Context, token string, ttl time.Duration) error {
	if ttl <= 0 {
		ttl = r.ttl
	}
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	key := "push:token:suppressed:" + token
	return r.client.SetEX(ctx, key, "1", ttl).Err()
}
