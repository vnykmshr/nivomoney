package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisConfig holds Redis connection configuration.
type RedisConfig struct {
	URL             string        // Redis connection URL (redis://:password@host:port/db)
	MaxRetries      int           // Maximum number of retries (default: 3)
	DialTimeout     time.Duration // Connection timeout (default: 5s)
	ReadTimeout     time.Duration // Read timeout (default: 3s)
	WriteTimeout    time.Duration // Write timeout (default: 3s)
	PoolSize        int           // Connection pool size (default: 10)
	MinIdleConns    int           // Minimum idle connections (default: 2)
	MaxIdleTime     time.Duration // Max idle time before closing (default: 5m)
	ConnMaxLifetime time.Duration // Max connection lifetime (default: 30m)
}

// DefaultRedisConfig returns a RedisConfig with sensible defaults.
func DefaultRedisConfig(url string) *RedisConfig {
	return &RedisConfig{
		URL:             url,
		MaxRetries:      3,
		DialTimeout:     5 * time.Second,
		ReadTimeout:     3 * time.Second,
		WriteTimeout:    3 * time.Second,
		PoolSize:        10,
		MinIdleConns:    2,
		MaxIdleTime:     5 * time.Minute,
		ConnMaxLifetime: 30 * time.Minute,
	}
}

// RedisCache implements Cache using Redis as the backend.
type RedisCache struct {
	client *redis.Client
}

// NewRedisCache creates a new Redis cache client.
// Returns an error if the connection cannot be established.
func NewRedisCache(cfg *RedisConfig) (*RedisCache, error) {
	opts, err := redis.ParseURL(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	// Apply configuration
	opts.MaxRetries = cfg.MaxRetries
	opts.DialTimeout = cfg.DialTimeout
	opts.ReadTimeout = cfg.ReadTimeout
	opts.WriteTimeout = cfg.WriteTimeout
	opts.PoolSize = cfg.PoolSize
	opts.MinIdleConns = cfg.MinIdleConns
	opts.ConnMaxIdleTime = cfg.MaxIdleTime
	opts.ConnMaxLifetime = cfg.ConnMaxLifetime

	client := redis.NewClient(opts)

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), cfg.DialTimeout)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisCache{client: client}, nil
}

// Get retrieves a value by key.
func (r *RedisCache) Get(ctx context.Context, key string) (string, bool, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("redis get error: %w", err)
	}
	return val, true, nil
}

// Set stores a value with the given key and TTL.
func (r *RedisCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	if err := r.client.Set(ctx, key, value, ttl).Err(); err != nil {
		return fmt.Errorf("redis set error: %w", err)
	}
	return nil
}

// Delete removes a value by key.
func (r *RedisCache) Delete(ctx context.Context, key string) error {
	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("redis delete error: %w", err)
	}
	return nil
}

// Exists checks if a key exists in the cache.
func (r *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	count, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("redis exists error: %w", err)
	}
	return count > 0, nil
}

// Ping checks the connection health.
func (r *RedisCache) Ping(ctx context.Context) error {
	if err := r.client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis ping error: %w", err)
	}
	return nil
}

// Close closes the Redis connection.
func (r *RedisCache) Close() error {
	return r.client.Close()
}

// Client returns the underlying Redis client for advanced operations.
func (r *RedisCache) Client() *redis.Client {
	return r.client
}
