package handler

import (
	"context"
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/vnykmshr/nivo/services/identity/internal/models"
	"github.com/vnykmshr/nivo/services/identity/internal/service"
	"github.com/vnykmshr/nivo/shared/errors"
	"github.com/vnykmshr/nivo/shared/response"
)

// ContextKey is a type for context keys to avoid collisions.
type ContextKey string

const (
	// UserContextKey is the key for storing user in context.
	UserContextKey ContextKey = "user"
	// PairedUserIDKey is the key for storing the paired user ID for User-Admin accounts.
	PairedUserIDKey ContextKey = "paired_user_id"
)

// AuthMiddleware provides authentication middleware functionality.
type AuthMiddleware struct {
	authService *service.AuthService
}

// NewAuthMiddleware creates a new authentication middleware.
func NewAuthMiddleware(authService *service.AuthService) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
	}
}

// Authenticate is a middleware that validates JWT tokens and sets user in context.
func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		token := extractBearerToken(r)
		if token == "" {
			response.Error(w, errors.Unauthorized("missing authorization token"))
			return
		}

		// Validate token and get user
		user, err := m.authService.ValidateToken(r.Context(), token)
		if err != nil {
			response.Error(w, err)
			return
		}

		// Set user in context
		ctx := context.WithValue(r.Context(), UserContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalAuthenticate is a middleware that validates JWT tokens if present, but doesn't require them.
func (m *AuthMiddleware) OptionalAuthenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		token := extractBearerToken(r)
		if token == "" {
			// No token, continue without authentication
			next.ServeHTTP(w, r)
			return
		}

		// Validate token and get user
		user, err := m.authService.ValidateToken(r.Context(), token)
		if err != nil {
			// Invalid token, continue without authentication
			next.ServeHTTP(w, r)
			return
		}

		// Set user in context
		ctx := context.WithValue(r.Context(), UserContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireStatus is a middleware that checks if the user has the required status.
func (m *AuthMiddleware) RequireStatus(statuses ...models.UserStatus) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get user from context
			user := getUserFromContext(r.Context())
			if user == nil {
				response.Error(w, errors.Unauthorized("user not authenticated"))
				return
			}

			// Check if user has required status
			for _, status := range statuses {
				if user.Status == status {
					next.ServeHTTP(w, r)
					return
				}
			}

			response.Error(w, errors.Forbidden("insufficient permissions"))
		})
	}
}

// RequireKYCVerified is a middleware that requires the user's KYC to be verified.
func (m *AuthMiddleware) RequireKYCVerified(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get user from context
		user := getUserFromContext(r.Context())
		if user == nil {
			response.Error(w, errors.Unauthorized("user not authenticated"))
			return
		}

		// Check if KYC is verified
		if user.KYC.Status != models.KYCStatusVerified {
			response.Error(w, errors.Forbidden("KYC verification required"))
			return
		}

		next.ServeHTTP(w, r)
	})
}

// getUserFromContext extracts the user from the request context (updated implementation).
func getUserFromContext(ctx context.Context) *models.User {
	user, ok := ctx.Value(UserContextKey).(*models.User)
	if !ok {
		return nil
	}
	return user
}

// JWTClaims represents the JWT token claims with RBAC support.
type JWTClaims struct {
	UserID      string   `json:"user_id"`
	Email       string   `json:"email"`
	Status      string   `json:"status"`
	Roles       []string `json:"roles,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
	jwt.RegisteredClaims
}

// extractPermissionsFromToken extracts permissions from JWT token without validating the signature.
// This is safe because the token signature has already been validated by the Authenticate middleware.
func extractPermissionsFromToken(tokenString string) ([]string, error) {
	// Parse token without validation (signature already validated)
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	token, _, err := parser.ParseUnverified(tokenString, &JWTClaims{})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok {
		return claims.Permissions, nil
	}

	return nil, fmt.Errorf("invalid token claims")
}

// extractRolesFromToken extracts roles from JWT token without validating the signature.
func extractRolesFromToken(tokenString string) ([]string, error) {
	// Parse token without validation (signature already validated)
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	token, _, err := parser.ParseUnverified(tokenString, &JWTClaims{})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok {
		return claims.Roles, nil
	}

	return nil, fmt.Errorf("invalid token claims")
}

// RequirePermission creates a middleware that checks if the user has the required permission.
// Must be chained after Authenticate middleware.
func (m *AuthMiddleware) RequirePermission(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token
			token := extractBearerToken(r)
			if token == "" {
				response.Error(w, errors.Unauthorized("missing authorization token"))
				return
			}

			// Extract permissions from token
			permissions, err := extractPermissionsFromToken(token)
			if err != nil {
				response.Error(w, errors.Forbidden("failed to extract permissions"))
				return
			}

			// Check if user has the required permission
			hasPermission := false
			for _, perm := range permissions {
				if perm == permission {
					hasPermission = true
					break
				}
			}

			if !hasPermission {
				response.Error(w, errors.Forbidden(fmt.Sprintf("missing required permission: %s", permission)))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAnyPermission creates a middleware that checks if the user has ANY of the required permissions.
func (m *AuthMiddleware) RequireAnyPermission(permissions ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token
			token := extractBearerToken(r)
			if token == "" {
				response.Error(w, errors.Unauthorized("missing authorization token"))
				return
			}

			// Extract permissions from token
			userPerms, err := extractPermissionsFromToken(token)
			if err != nil {
				response.Error(w, errors.Forbidden("failed to extract permissions"))
				return
			}

			// Check if user has any of the required permissions
			hasPermission := false
			for _, requiredPerm := range permissions {
				for _, userPerm := range userPerms {
					if userPerm == requiredPerm {
						hasPermission = true
						break
					}
				}
				if hasPermission {
					break
				}
			}

			if !hasPermission {
				response.Error(w, errors.Forbidden(fmt.Sprintf("missing any of required permissions: %v", permissions)))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireRole creates a middleware that checks if the user has the required role.
func (m *AuthMiddleware) RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token
			token := extractBearerToken(r)
			if token == "" {
				response.Error(w, errors.Unauthorized("missing authorization token"))
				return
			}

			// Extract roles from token
			roles, err := extractRolesFromToken(token)
			if err != nil {
				response.Error(w, errors.Forbidden("failed to extract roles"))
				return
			}

			// Check if user has the required role
			hasRole := false
			for _, r := range roles {
				if r == role {
					hasRole = true
					break
				}
			}

			if !hasRole {
				response.Error(w, errors.Forbidden(fmt.Sprintf("missing required role: %s", role)))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAnyRole creates a middleware that checks if the user has ANY of the required roles.
func (m *AuthMiddleware) RequireAnyRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token
			token := extractBearerToken(r)
			if token == "" {
				response.Error(w, errors.Unauthorized("missing authorization token"))
				return
			}

			// Extract roles from token
			userRoles, err := extractRolesFromToken(token)
			if err != nil {
				response.Error(w, errors.Forbidden("failed to extract roles"))
				return
			}

			// Check if user has any of the required roles
			hasRole := false
			for _, requiredRole := range roles {
				for _, userRole := range userRoles {
					if userRole == requiredRole {
						hasRole = true
						break
					}
				}
				if hasRole {
					break
				}
			}

			if !hasRole {
				response.Error(w, errors.Forbidden(fmt.Sprintf("missing any of required roles: %v", roles)))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// UserAdminValidation is a middleware that validates User-Admin access scope.
// For User-Admin accounts (account_type = 'user_admin'), this middleware:
// 1. Loads the paired regular user ID into context
// 2. Validates that requests targeting a user ID are scoped to the paired user
type UserAdminValidation struct {
	authService *service.AuthService
}

// NewUserAdminValidation creates a new User-Admin validation middleware.
func NewUserAdminValidation(authService *service.AuthService) *UserAdminValidation {
	return &UserAdminValidation{
		authService: authService,
	}
}

// ValidatePairing ensures User-Admin accounts can only access their paired user's data.
// This middleware must be chained after Authenticate middleware.
// It extracts target user ID from path (userId parameter) and validates pairing.
func (v *UserAdminValidation) ValidatePairing(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get user from context
		user := getUserFromContext(r.Context())
		if user == nil {
			response.Error(w, errors.Unauthorized("user not authenticated"))
			return
		}

		// If not a User-Admin, continue without pairing validation
		if user.AccountType != models.AccountTypeUserAdmin {
			next.ServeHTTP(w, r)
			return
		}

		// User-Admin: Load paired user ID into context
		pairedUserID, err := v.authService.GetPairedUserID(r.Context(), user.ID)
		if err != nil {
			response.Error(w, errors.Forbidden("failed to validate user-admin pairing"))
			return
		}

		// Add paired user ID to context
		ctx := context.WithValue(r.Context(), PairedUserIDKey, pairedUserID)

		// Check if there's a target user ID in the path
		targetUserID := r.PathValue("userId")
		if targetUserID != "" && targetUserID != pairedUserID {
			// User-Admin is trying to access a different user's data
			response.Error(w, errors.Forbidden("access denied: user-admin can only access paired user data"))
			return
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// LoadPairedUserID loads the paired user ID into context without validating access.
// Use this for endpoints where User-Admin should be aware of their paired user.
func (v *UserAdminValidation) LoadPairedUserID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get user from context
		user := getUserFromContext(r.Context())
		if user == nil || user.AccountType != models.AccountTypeUserAdmin {
			next.ServeHTTP(w, r)
			return
		}

		// User-Admin: Load paired user ID into context
		pairedUserID, err := v.authService.GetPairedUserID(r.Context(), user.ID)
		if err == nil {
			ctx := context.WithValue(r.Context(), PairedUserIDKey, pairedUserID)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		next.ServeHTTP(w, r)
	})
}

// GetPairedUserIDFromContext extracts the paired user ID from context.
// Returns empty string if not a User-Admin or pairing not loaded.
func GetPairedUserIDFromContext(ctx context.Context) string {
	pairedUserID, ok := ctx.Value(PairedUserIDKey).(string)
	if !ok {
		return ""
	}
	return pairedUserID
}

// GetTargetUserID returns the appropriate user ID based on account type.
// For User-Admin accounts, returns the paired user ID.
// For regular users, returns the authenticated user's ID.
func GetTargetUserID(ctx context.Context) string {
	user := getUserFromContext(ctx)
	if user == nil {
		return ""
	}

	// For User-Admin, use paired user ID
	if user.AccountType == models.AccountTypeUserAdmin {
		pairedUserID := GetPairedUserIDFromContext(ctx)
		if pairedUserID != "" {
			return pairedUserID
		}
	}

	// For regular users, use their own ID
	return user.ID
}
