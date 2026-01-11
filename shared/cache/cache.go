// Package cache provides a cache interface and implementations for caching data.
package cache

import (
	"context"
	"time"
)

// Cache defines the interface for cache operations.
// All operations should be context-aware and support graceful degradation.
type Cache interface {
	// Get retrieves a value by key. Returns empty string and false if not found.
	Get(ctx context.Context, key string) (string, bool, error)

	// Set stores a value with the given key and TTL.
	Set(ctx context.Context, key string, value string, ttl time.Duration) error

	// Delete removes a value by key.
	Delete(ctx context.Context, key string) error

	// Exists checks if a key exists in the cache.
	Exists(ctx context.Context, key string) (bool, error)

	// Ping checks the connection health.
	Ping(ctx context.Context) error

	// Close closes the cache connection.
	Close() error
}

// NoOpCache is a cache implementation that does nothing.
// Useful when caching is disabled or as a fallback.
type NoOpCache struct{}

func (n *NoOpCache) Get(ctx context.Context, key string) (string, bool, error) {
	return "", false, nil
}

func (n *NoOpCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	return nil
}

func (n *NoOpCache) Delete(ctx context.Context, key string) error {
	return nil
}

func (n *NoOpCache) Exists(ctx context.Context, key string) (bool, error) {
	return false, nil
}

func (n *NoOpCache) Ping(ctx context.Context) error {
	return nil
}

func (n *NoOpCache) Close() error {
	return nil
}

// NewNoOpCache creates a new no-op cache.
func NewNoOpCache() *NoOpCache {
	return &NoOpCache{}
}
