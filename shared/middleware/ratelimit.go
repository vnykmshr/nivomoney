package middleware

import (
	"fmt"
	"net/http"
	"strings"
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

	// TrustProxyHeaders enables trusting X-Forwarded-For and X-Real-IP headers.
	// SECURITY: Only enable this when behind a trusted reverse proxy (nginx, traefik, etc.)
	// that properly sets these headers. Leaving this false prevents IP spoofing attacks.
	TrustProxyHeaders bool
}

// DefaultRateLimitConfig returns a sensible default configuration
// Suitable for authentication endpoints (prevent brute force)
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		RequestsPerMinute: 60, // 60 requests per minute (1 per second)
		BurstSize:         20, // Allow burst of 20 requests
		CleanupInterval:   10 * time.Minute,
		TrustProxyHeaders: false, // Secure default: don't trust proxy headers
	}
}

// StrictRateLimitConfig returns a stricter configuration
// Suitable for sensitive operations (KYC verification, money transfers)
func StrictRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		RequestsPerMinute: 30, // 30 requests per minute
		BurstSize:         10, // Allow burst of 10 requests
		CleanupInterval:   10 * time.Minute,
		TrustProxyHeaders: false, // Secure default: don't trust proxy headers
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
			// Get client IP (use proxy headers only if explicitly trusted)
			var ip string
			if config.TrustProxyHeaders {
				ip = getClientIPFromProxy(r)
			} else {
				ip = getClientIP(r)
			}

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
				_, _ = w.Write([]byte(`{"success":false,"error":"rate limit exceeded","message":"Too many requests. Please try again later."}`))
				return
			}

			// Set rate limit headers for successful requests
			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", config.RequestsPerMinute))
			w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%.0f", tokens))

			next.ServeHTTP(w, r)
		})
	}
}

// getClientIP extracts the client IP address from the request.
// Uses RemoteAddr by default for security. Proxy headers are only trusted
// when TrustProxyHeaders is true (should only be enabled behind a trusted proxy).
//
// SECURITY: X-Forwarded-For and X-Real-IP headers can be spoofed by clients.
// Only enable TrustProxyHeaders when your server is behind a trusted reverse proxy.
func getClientIP(r *http.Request) string {
	// Extract IP from RemoteAddr (format: "ip:port" or "ip")
	ip := extractIP(r.RemoteAddr)
	if ip == "" {
		ip = r.RemoteAddr
	}
	return ip
}

// getClientIPFromProxy extracts client IP from proxy headers.
// Only use this behind a trusted reverse proxy.
func getClientIPFromProxy(r *http.Request) string {
	// Check X-Forwarded-For header (most common)
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		// X-Forwarded-For format: "client, proxy1, proxy2"
		// Take the first IP (the original client)
		if idx := strings.Index(forwarded, ","); idx != -1 {
			return strings.TrimSpace(forwarded[:idx])
		}
		return strings.TrimSpace(forwarded)
	}

	// Check X-Real-IP header (used by nginx)
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return strings.TrimSpace(realIP)
	}

	// Fallback to RemoteAddr
	return getClientIP(r)
}

// extractIP extracts just the IP address from a host:port string
func extractIP(addr string) string {
	// Handle IPv6 format: [::1]:8080
	if len(addr) > 0 && addr[0] == '[' {
		if idx := strings.Index(addr, "]"); idx != -1 {
			return addr[1:idx]
		}
	}
	// Handle IPv4 format: 127.0.0.1:8080
	if idx := strings.LastIndex(addr, ":"); idx != -1 {
		// Check if there's only one colon (IPv4) or if it's after ] (IPv6)
		if strings.Count(addr, ":") == 1 {
			return addr[:idx]
		}
	}
	return addr
}
