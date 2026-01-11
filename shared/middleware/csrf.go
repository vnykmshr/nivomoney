package middleware

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/vnykmshr/nivo/shared/errors"
	"github.com/vnykmshr/nivo/shared/response"
)

const (
	// CSRFTokenHeader is the header name for CSRF token.
	CSRFTokenHeader = "X-CSRF-Token" //nolint:gosec // Not a credential, just a header name
	// CSRFCookieName is the cookie name for CSRF token.
	CSRFCookieName = "csrf_token"
	// CSRFTokenLength is the length of the CSRF token in bytes (32 bytes = 64 hex chars).
	CSRFTokenLength = 32
)

// CSRFConfig holds configuration for CSRF middleware.
type CSRFConfig struct {
	// SkipPaths are paths that don't require CSRF validation (e.g., login, register).
	SkipPaths []string
	// CookiePath is the path for the CSRF cookie. Default is "/".
	CookiePath string
	// CookieSecure sets the Secure flag on the cookie. Should be true in production.
	CookieSecure bool
	// CookieSameSite sets the SameSite attribute. Default is SameSiteLaxMode.
	CookieSameSite http.SameSite
}

// DefaultCSRFConfig returns a default CSRF configuration.
func DefaultCSRFConfig() CSRFConfig {
	return CSRFConfig{
		SkipPaths: []string{
			"/api/v1/auth/login",
			"/api/v1/auth/register",
			"/api/v1/auth/refresh",
			"/health",
			"/metrics",
		},
		CookiePath:     "/",
		CookieSecure:   false, // Set to true in production via config
		CookieSameSite: http.SameSiteLaxMode,
	}
}

// CSRF creates a middleware that implements CSRF protection using double-submit cookie pattern.
//
// How it works:
// 1. On any request, if no CSRF cookie exists, generate one and set it
// 2. On mutating requests (POST, PUT, DELETE, PATCH), validate that X-CSRF-Token header matches cookie
// 3. Frontend reads the csrf_token cookie and sends it in X-CSRF-Token header
//
// This is a stateless approach that doesn't require server-side session storage.
func CSRF(config CSRFConfig) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if path should skip CSRF validation
			for _, path := range config.SkipPaths {
				if strings.HasPrefix(r.URL.Path, path) {
					next.ServeHTTP(w, r)
					return
				}
			}

			// Get or generate CSRF token
			cookie, err := r.Cookie(CSRFCookieName)
			var csrfToken string

			if err == http.ErrNoCookie || cookie == nil || cookie.Value == "" {
				// Generate new token
				csrfToken, err = generateCSRFToken()
				if err != nil {
					response.Error(w, errors.Internal("failed to generate CSRF token"))
					return
				}

				// Set cookie for future requests
				http.SetCookie(w, &http.Cookie{
					Name:     CSRFCookieName,
					Value:    csrfToken,
					Path:     config.CookiePath,
					Secure:   config.CookieSecure,
					HttpOnly: false, // Must be readable by JavaScript
					SameSite: config.CookieSameSite,
				})
			} else {
				csrfToken = cookie.Value
			}

			// For safe methods (GET, HEAD, OPTIONS), just continue
			if isSafeMethod(r.Method) {
				next.ServeHTTP(w, r)
				return
			}

			// For mutating methods, validate CSRF token
			headerToken := r.Header.Get(CSRFTokenHeader)
			if headerToken == "" {
				response.Error(w, errors.Forbidden("missing CSRF token"))
				return
			}

			// Compare tokens (constant-time comparison to prevent timing attacks)
			if !secureCompare(headerToken, csrfToken) {
				response.Error(w, errors.Forbidden("invalid CSRF token"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// generateCSRFToken generates a cryptographically secure random token.
func generateCSRFToken() (string, error) {
	bytes := make([]byte, CSRFTokenLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// isSafeMethod returns true if the HTTP method is considered safe (doesn't modify state).
func isSafeMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
		return true
	default:
		return false
	}
}

// secureCompare performs a constant-time comparison of two strings.
// Uses crypto/subtle to prevent timing attacks when comparing CSRF tokens.
func secureCompare(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// CSRFTokenFromRequest extracts the CSRF token from the request cookie.
// This can be used by handlers that need to include the token in responses.
func CSRFTokenFromRequest(r *http.Request) string {
	cookie, err := r.Cookie(CSRFCookieName)
	if err != nil || cookie == nil {
		return ""
	}
	return cookie.Value
}
