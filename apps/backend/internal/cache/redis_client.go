package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// CacheClientInterface defines the contract for cache operations.
type CacheClientInterface interface {
	SetEstimation(ctx context.Context, key string, result interface{}, ttl time.Duration) error
	GetEstimation(ctx context.Context, key string, result interface{}) error
	Delete(ctx context.Context, key string) error
	Ping(ctx context.Context) error
}

// CacheClient implements Redis cache operations.
type CacheClient struct {
	client *redis.Client
}

// NewCacheClient creates a new cache client.
func NewCacheClient(client *redis.Client) *CacheClient {
	return &CacheClient{
		client: client,
	}
}

// SetEstimation stores an estimation result in cache.
func (c *CacheClient) SetEstimation(ctx context.Context, key string, result interface{}, ttl time.Duration) error {
	data, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal estimation result: %w", err)
	}

	return c.client.Set(ctx, key, data, ttl).Err()
}

// GetEstimation retrieves an estimation result from cache.
func (c *CacheClient) GetEstimation(ctx context.Context, key string, result interface{}) error {
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("cache miss")
		}
		return fmt.Errorf("failed to get from cache: %w", err)
	}

	if err := json.Unmarshal(data, result); err != nil {
		return fmt.Errorf("failed to unmarshal estimation result: %w", err)
	}

	return nil
}

// Delete removes a key from cache.
func (c *CacheClient) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

// Ping checks Redis connectivity.
func (c *CacheClient) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

