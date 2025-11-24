package proxy

import (
	"fmt"
	"os"
)

// ServiceRegistry holds the URLs of all backend services.
type ServiceRegistry struct {
	Identity    string
	Ledger      string
	RBAC        string
	Transaction string
	Wallet      string
	Risk        string
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
	}
}

// GetServiceURL returns the URL for a given service name.
func (r *ServiceRegistry) GetServiceURL(serviceName string) (string, error) {
	switch serviceName {
	case "identity":
		return r.Identity, nil
	case "ledger":
		return r.Ledger, nil
	case "rbac":
		return r.RBAC, nil
	case "transaction":
		return r.Transaction, nil
	case "wallet":
		return r.Wallet, nil
	case "risk":
		return r.Risk, nil
	default:
		return "", fmt.Errorf("unknown service: %s", serviceName)
	}
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
	}
}

// getEnvOrDefault returns the value of an environment variable or a default value.
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
