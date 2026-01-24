package proxy

import (
	"fmt"
	"os"
	"regexp"
)

// ServiceRegistry holds the URLs of all backend services.
type ServiceRegistry struct {
	Identity    string
	Ledger      string
	RBAC        string
	Transaction string
	Wallet      string
	Risk        string
	Simulation  string
}

// NewServiceRegistry creates a new service registry from environment variables.
func NewServiceRegistry() *ServiceRegistry {
	return &ServiceRegistry{
		Identity:    getEnvOrDefault("IDENTITY_SERVICE_URL", "http://identity-service:8080"),
		Ledger:      getEnvOrDefault("LEDGER_SERVICE_URL", "http://ledger-service:8081"),
		RBAC:        getEnvOrDefault("RBAC_SERVICE_URL", "http://rbac-service:8082"),
		Transaction: getEnvOrDefault("TRANSACTION_SERVICE_URL", "http://transaction-service:8084"),
		Wallet:      getEnvOrDefault("WALLET_SERVICE_URL", "http://wallet-service:8083"),
		Risk:        getEnvOrDefault("RISK_SERVICE_URL", "http://risk-service:8085"),
		Simulation:  getEnvOrDefault("SIMULATION_SERVICE_URL", "http://simulation-service:8086"),
	}
}

// ServiceInfo contains URL and routing configuration for a service.
type ServiceInfo struct {
	URL     string
	IsAlias bool // If true, don't strip the service name from path
}

// GetServiceInfo returns the service configuration for a given service name.
func (r *ServiceRegistry) GetServiceInfo(serviceName string) (*ServiceInfo, error) {
	switch serviceName {
	case "identity":
		return &ServiceInfo{URL: r.Identity, IsAlias: false}, nil
	case "auth", "users", "verifications", "user-admin":
		// "auth", "users", "verifications", "user-admin" are aliases - preserve path segment
		return &ServiceInfo{URL: r.Identity, IsAlias: true}, nil
	case "ledger":
		return &ServiceInfo{URL: r.Ledger, IsAlias: false}, nil
	case "rbac":
		return &ServiceInfo{URL: r.RBAC, IsAlias: false}, nil
	case "transaction":
		return &ServiceInfo{URL: r.Transaction, IsAlias: false}, nil
	case "transactions":
		// "transactions" is alias - preserve path segment
		return &ServiceInfo{URL: r.Transaction, IsAlias: true}, nil
	case "wallet":
		return &ServiceInfo{URL: r.Wallet, IsAlias: false}, nil
	case "wallets":
		// "wallets" is alias - preserve path segment
		return &ServiceInfo{URL: r.Wallet, IsAlias: true}, nil
	case "risk":
		return &ServiceInfo{URL: r.Risk, IsAlias: false}, nil
	case "simulation":
		return &ServiceInfo{URL: r.Simulation, IsAlias: false}, nil
	default:
		return nil, fmt.Errorf("unknown service: %s", serviceName)
	}
}

// GetServiceURL returns the URL for a given service name (for backward compatibility).
func (r *ServiceRegistry) GetServiceURL(serviceName string) (string, error) {
	info, err := r.GetServiceInfo(serviceName)
	if err != nil {
		return "", err
	}
	return info.URL, nil
}

// AllServices returns a map of all registered services.
func (r *ServiceRegistry) AllServices() map[string]string {
	return map[string]string{
		"identity":    r.Identity,
		"ledger":      r.Ledger,
		"rbac":        r.RBAC,
		"transaction": r.Transaction,
		"wallet":      r.Wallet,
		"risk":        r.Risk,
		"simulation":  r.Simulation,
	}
}

// pathRoutingRule defines a special path pattern that routes to a specific service.
type pathRoutingRule struct {
	pattern *regexp.Regexp
	service string
}

// pathRoutingRules defines paths that need special routing (checked before default segment matching).
// These handle cases where nested resources belong to a different service than the parent.
var pathRoutingRules = []pathRoutingRule{
	// Wallet-related transaction endpoints belong to transaction service
	{pattern: regexp.MustCompile(`^wallets/[^/]+/transactions$`), service: "transactions"},
	{pattern: regexp.MustCompile(`^wallets/[^/]+/spending-summary$`), service: "transactions"},
	{pattern: regexp.MustCompile(`^wallets/[^/]+/statements/`), service: "transactions"},
}

// GetServiceByPath checks if the path matches any special routing rules.
// Returns service info if a rule matches, or nil if default routing should be used.
func (r *ServiceRegistry) GetServiceByPath(path string) *ServiceInfo {
	for _, rule := range pathRoutingRules {
		if rule.pattern.MatchString(path) {
			info, err := r.GetServiceInfo(rule.service)
			if err == nil {
				return info
			}
		}
	}
	return nil
}

// getEnvOrDefault returns the value of an environment variable or a default value.
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
