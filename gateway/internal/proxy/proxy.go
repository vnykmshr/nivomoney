package proxy

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/vnykmshr/nivo/shared/errors"
	"github.com/vnykmshr/nivo/shared/response"
)

// Gateway handles proxying requests to backend services.
type Gateway struct {
	registry *ServiceRegistry
}

// NewGateway creates a new API gateway.
func NewGateway(registry *ServiceRegistry) *Gateway {
	return &Gateway{
		registry: registry,
	}
}

// ProxyRequest proxies the request to the appropriate backend service.
func (g *Gateway) ProxyRequest(w http.ResponseWriter, r *http.Request) {
	// Extract service name from path: /api/v1/{service}/...
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/")
	parts := strings.SplitN(path, "/", 2)

	if len(parts) < 1 {
		response.Error(w, errors.BadRequest("invalid request path"))
		return
	}

	serviceName := parts[0]

	// Get backend service info
	serviceInfo, err := g.registry.GetServiceInfo(serviceName)
	if err != nil {
		response.Error(w, errors.NotFound(err.Error()))
		return
	}

	// Parse backend URL
	target, err := url.Parse(serviceInfo.URL)
	if err != nil {
		log.Printf("[gateway] Failed to parse backend URL %s: %v", serviceInfo.URL, err)
		response.Error(w, errors.Internal("failed to parse backend URL"))
		return
	}

	// Create reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(target)

	// Customize the director to strip service name from path (unless it's an alias)
	originalDirector := proxy.Director
	isAlias := serviceInfo.IsAlias
	proxy.Director = func(req *http.Request) {
		originalDirector(req)

		if isAlias {
			// For aliases (auth, users, wallets, transactions), preserve the full path
			// Example: /api/v1/auth/login -> /api/v1/auth/login (unchanged)
			req.URL.Path = r.URL.Path
		} else {
			// Strip service name from path: /api/v1/{service}/... -> /api/v1/...
			// Example: /api/v1/identity/auth/register -> /api/v1/auth/register
			pathWithoutPrefix := strings.TrimPrefix(r.URL.Path, "/api/v1/")
			pathParts := strings.SplitN(pathWithoutPrefix, "/", 2)

			if len(pathParts) > 1 {
				// Reconstruct path without service name
				req.URL.Path = "/api/v1/" + pathParts[1]
			} else {
				// If no path after service name, just use /api/v1/
				req.URL.Path = "/api/v1/"
			}
		}

		req.URL.RawQuery = r.URL.RawQuery

		// Set X-Forwarded headers
		req.Header.Set("X-Forwarded-Host", r.Host)
		req.Header.Set("X-Forwarded-Proto", getScheme(r))
		req.Header.Set("X-Real-IP", getClientIP(r))

		// Add request ID if present
		if reqID := r.Header.Get("X-Request-ID"); reqID != "" {
			req.Header.Set("X-Request-ID", reqID)
		}

		log.Printf("[gateway] Proxying %s %s → %s%s", r.Method, r.URL.Path, target.Host, req.URL.Path)
	}

	// Customize error handler
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("[gateway] Proxy error for %s: %v", r.URL.Path, err)
		response.Error(w, errors.Unavailable("backend service unavailable"))
	}

	// Proxy the request
	proxy.ServeHTTP(w, r)
}

// getScheme returns the request scheme (http or https).
func getScheme(r *http.Request) string {
	if r.TLS != nil {
		return "https"
	}
	if scheme := r.Header.Get("X-Forwarded-Proto"); scheme != "" {
		return scheme
	}
	return "http"
}

// getClientIP returns the client IP address.
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if colonIndex := strings.LastIndex(ip, ":"); colonIndex != -1 {
		ip = ip[:colonIndex]
	}
	return ip
}
