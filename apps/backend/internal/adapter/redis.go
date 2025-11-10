package adapter

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisDB represents the Redis connection client.
type RedisDB struct {
	Client *redis.Client
}

// RedisConfig holds configuration for Redis connection.
type RedisConfig struct {
	URL     string
	Enabled bool
	DB      int
}

// NewRedis creates a new Redis client.
func NewRedis(cfg RedisConfig) (*RedisDB, error) {
	if !cfg.Enabled {
		return nil, fmt.Errorf("redis is disabled")
	}

	opt, err := redis.ParseURL(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis URL: %w", err)
	}

	// Override DB if specified
	if cfg.DB > 0 {
		opt.DB = cfg.DB
	}

	client := redis.NewClient(opt)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &RedisDB{
		Client: client,
	}, nil
}

