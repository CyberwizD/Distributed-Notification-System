package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisRepository is a repository for Redis.
type RedisRepository struct {
	Client *redis.Client
}

// NewRedisRepository creates a new RedisRepository.
func NewRedisRepository(client *redis.Client) *RedisRepository {
	return &RedisRepository{Client: client}
}

// GetJSON fetches a JSON payload from Redis and unmarshals it into dest.
func (r *RedisRepository) GetJSON(ctx context.Context, key string, dest interface{}) (bool, error) {
	if r == nil || r.Client == nil {
		return false, nil
	}
	data, err := r.Client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if err := json.Unmarshal(data, dest); err != nil {
		return false, err
	}
	return true, nil
}

// SetJSON stores a JSON payload in Redis with the provided TTL.
func (r *RedisRepository) SetJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if r == nil || r.Client == nil {
		return nil
	}
	if ttl <= 0 {
		return nil
	}
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return r.Client.SetEX(ctx, key, data, ttl).Err()
}
