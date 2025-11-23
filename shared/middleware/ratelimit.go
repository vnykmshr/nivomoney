package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

// RateLimitConfig holds configuration for rate limiting
type RateLimitConfig struct {
	// RequestsPerMinute is the maximum number of requests allowed per minute
	RequestsPerMinute int

	// BurstSize is the maximum burst of requests allowed
	BurstSize int

	// CleanupInterval is how often to clean up expired entries (default: 10 minutes)
	CleanupInterval time.Duration
}

// DefaultRateLimitConfig returns a sensible default configuration
// Suitable for authentication endpoints (prevent brute force)
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		RequestsPerMinute: 20, // 20 requests per minute
		BurstSize:         5,  // Allow burst of 5 requests
		CleanupInterval:   10 * time.Minute,
	}
}

// StrictRateLimitConfig returns a stricter configuration
// Suitable for sensitive operations (KYC verification, money transfers)
func StrictRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		RequestsPerMinute: 5, // 5 requests per minute
		BurstSize:         2, // Allow burst of 2 requests
		CleanupInterval:   10 * time.Minute,
	}
}

// visitor tracks rate limit state for a single IP address
type visitor struct {
	tokens     float64   // Current number of tokens
	lastUpdate time.Time // Last time tokens were updated
	mu         sync.Mutex
}

// rateLimiter implements token bucket rate limiting
type rateLimiter struct {
	visitors        sync.Map // map[string]*visitor (IP -> visitor)
	config          RateLimitConfig
	tokensPerSecond float64 // Calculated from RequestsPerMinute
	cleanupTicker   *time.Ticker
	stopCleanup     chan struct{}
}

// newRateLimiter creates a new rate limiter with the given configuration
func newRateLimiter(config RateLimitConfig) *rateLimiter {
	rl := &rateLimiter{
		config:          config,
		tokensPerSecond: float64(config.RequestsPerMinute) / 60.0,
		stopCleanup:     make(chan struct{}),
	}

	// Start cleanup goroutine
	if config.CleanupInterval > 0 {
		rl.cleanupTicker = time.NewTicker(config.CleanupInterval)
		go rl.cleanupLoop()
	}

	return rl
}

// cleanupLoop periodically removes old visitor entries
func (rl *rateLimiter) cleanupLoop() {
	for {
		select {
		case <-rl.cleanupTicker.C:
			rl.cleanup()
		case <-rl.stopCleanup:
			rl.cleanupTicker.Stop()
			return
		}
	}
}

// cleanup removes visitor entries that haven't been accessed recently
func (rl *rateLimiter) cleanup() {
	cutoff := time.Now().Add(-rl.config.CleanupInterval)
	rl.visitors.Range(func(key, value interface{}) bool {
		v := value.(*visitor)
		v.mu.Lock()
		if v.lastUpdate.Before(cutoff) {
			rl.visitors.Delete(key)
		}
		v.mu.Unlock()
		return true
	})
}

// getVisitor gets or creates a visitor for the given IP
func (rl *rateLimiter) getVisitor(ip string) *visitor {
	v, ok := rl.visitors.Load(ip)
	if !ok {
		// Create new visitor with full token bucket
		newVisitor := &visitor{
			tokens:     float64(rl.config.BurstSize),
			lastUpdate: time.Now(),
		}
		rl.visitors.Store(ip, newVisitor)
		return newVisitor
	}
	return v.(*visitor)
}

// allow checks if a request from the given IP should be allowed
func (rl *rateLimiter) allow(ip string) (bool, float64) {
	v := rl.getVisitor(ip)

	v.mu.Lock()
	defer v.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(v.lastUpdate).Seconds()

	// Refill tokens based on elapsed time
	v.tokens += elapsed * rl.tokensPerSecond

	// Cap at burst size
	if v.tokens > float64(rl.config.BurstSize) {
		v.tokens = float64(rl.config.BurstSize)
	}

	v.lastUpdate = now

	// Check if we have at least 1 token
	if v.tokens >= 1.0 {
		v.tokens--
		return true, v.tokens
	}

	return false, v.tokens
}

// RateLimit creates a rate limiting middleware with the given configuration
func RateLimit(config RateLimitConfig) func(http.Handler) http.Handler {
	limiter := newRateLimiter(config)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get client IP
			ip := getClientIP(r)

			// Check rate limit
			allowed, tokens := limiter.allow(ip)

			if !allowed {
				// Calculate retry-after in seconds
				retryAfter := int((1.0 - tokens) / limiter.tokensPerSecond)
				if retryAfter < 1 {
					retryAfter = 1
				}

				// Set rate limit headers
				w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", config.RequestsPerMinute))
				w.Header().Set("X-RateLimit-Remaining", "0")
				w.Header().Set("Retry-After", fmt.Sprintf("%d", retryAfter))

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"success":false,"error":"rate limit exceeded","message":"Too many requests. Please try again later."}`))
				return
			}

			// Set rate limit headers for successful requests
			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", config.RequestsPerMinute))
			w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%.0f", tokens))

			next.ServeHTTP(w, r)
		})
	}
}

// getClientIP extracts the client IP address from the request
// Checks X-Forwarded-For, X-Real-IP, and RemoteAddr in that order
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (most common in production behind proxies)
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		// Format: "client, proxy1, proxy2"
		// Note: In production, you should validate this comes from trusted proxy
		return forwarded
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fallback to RemoteAddr
	return r.RemoteAddr
}
