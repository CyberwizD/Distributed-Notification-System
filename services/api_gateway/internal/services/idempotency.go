package services

import (
	"context"
	"time"

	"github.com/CyberwizD/Distributed-Notification-System/services/api_gateway/internal/repository"
)

const idempotencyKeyTTL = 24 * time.Hour

// IdempotencyService handles idempotency checks.
type IdempotencyService struct {
	redisRepo *repository.RedisRepository
}

// NewIdempotencyService creates a new IdempotencyService.
func NewIdempotencyService(redisRepo *repository.RedisRepository) *IdempotencyService {
	return &IdempotencyService{redisRepo: redisRepo}
}

// IsDuplicate checks if a request ID has been seen before.
func (s *IdempotencyService) IsDuplicate(requestID string) (bool, error) {
	key := "idempotency:" + requestID
	// Use SetNX to atomically set the key if it doesn't exist.
	// If the key already exists, SetNX returns false.
	wasSet, err := s.redisRepo.Client.SetNX(context.Background(), key, "processed", idempotencyKeyTTL).Result()
	if err != nil {
		return false, err
	}
	// If the key was set, it's a new request. If not, it's a duplicate.
	return !wasSet, nil
}
