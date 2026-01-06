package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/vnykmshr/nivo/shared/errors"
	"github.com/vnykmshr/nivo/shared/response"
)

// ContextKey is a custom type for context keys to avoid collisions.
type ContextKey string

const (
	// UserIDKey is the context key for user ID.
	UserIDKey ContextKey = "user_id"
	// UserEmailKey is the context key for user email.
	UserEmailKey ContextKey = "user_email"
	// UserStatusKey is the context key for user status.
	UserStatusKey ContextKey = "user_status"
	// UserRolesKey is the context key for user roles.
	UserRolesKey ContextKey = "user_roles"
	// UserPermissionsKey is the context key for user permissions.
	UserPermissionsKey ContextKey = "user_permissions"
	// JWTTokenKey is the context key for the JWT token string (for service-to-service forwarding).
	JWTTokenKey ContextKey = "jwt_token"
	// AccountTypeKey is the context key for account type (user, user_admin).
	AccountTypeKey ContextKey = "account_type"
)

// JWTClaims represents the JWT token claims structure.
type JWTClaims struct {
	UserID      string   `json:"user_id"`
	Email       string   `json:"email"`
	Status      string   `json:"status"`
	AccountType string   `json:"account_type,omitempty"`
	Roles       []string `json:"roles,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
	jwt.RegisteredClaims
}

// AuthConfig holds configuration for auth middleware.
type AuthConfig struct {
	JWTSecret string
	// Optional: Skip auth for certain paths
	SkipPaths []string
}

// Auth creates a middleware that validates JWT tokens and extracts user claims.
func Auth(config AuthConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if path should skip auth
			for _, path := range config.SkipPaths {
				if strings.HasPrefix(r.URL.Path, path) {
					next.ServeHTTP(w, r)
					return
				}
			}

			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.Error(w, errors.Unauthorized("missing authorization header"))
				return
			}

			// Check for Bearer prefix
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				response.Error(w, errors.Unauthorized("invalid authorization header format"))
				return
			}

			tokenString := parts[1]

			// Parse and validate JWT
			claims := &JWTClaims{}
			token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
				// Validate signing method
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(config.JWTSecret), nil
			})

			if err != nil || !token.Valid {
				response.Error(w, errors.Unauthorized("invalid or expired token"))
				return
			}

			// Add claims and token to request context
			ctx := r.Context()
			ctx = context.WithValue(ctx, UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, UserEmailKey, claims.Email)
			ctx = context.WithValue(ctx, UserStatusKey, claims.Status)
			ctx = context.WithValue(ctx, AccountTypeKey, claims.AccountType)
			ctx = context.WithValue(ctx, UserRolesKey, claims.Roles)
			ctx = context.WithValue(ctx, UserPermissionsKey, claims.Permissions)
			ctx = context.WithValue(ctx, JWTTokenKey, tokenString) // Store token for service-to-service forwarding

			// Continue with updated context
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequirePermission creates a middleware that checks if the user has the required permission.
func RequirePermission(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get permissions from context
			permissions, ok := r.Context().Value(UserPermissionsKey).([]string)
			if !ok {
				response.Error(w, errors.Forbidden("no permissions found in token"))
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
func RequireAnyPermission(permissions ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get permissions from context
			userPerms, ok := r.Context().Value(UserPermissionsKey).([]string)
			if !ok {
				response.Error(w, errors.Forbidden("no permissions found in token"))
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
func RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get roles from context
			roles, ok := r.Context().Value(UserRolesKey).([]string)
			if !ok {
				response.Error(w, errors.Forbidden("no roles found in token"))
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
func RequireAnyRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get roles from context
			userRoles, ok := r.Context().Value(UserRolesKey).([]string)
			if !ok {
				response.Error(w, errors.Forbidden("no roles found in token"))
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

// GetUserID extracts the user ID from the request context.
func GetUserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey).(string)
	return userID, ok
}

// GetUserEmail extracts the user email from the request context.
func GetUserEmail(ctx context.Context) (string, bool) {
	email, ok := ctx.Value(UserEmailKey).(string)
	return email, ok
}

// GetUserStatus extracts the user status from the request context.
func GetUserStatus(ctx context.Context) (string, bool) {
	status, ok := ctx.Value(UserStatusKey).(string)
	return status, ok
}

// GetUserRoles extracts the user roles from the request context.
func GetUserRoles(ctx context.Context) ([]string, bool) {
	roles, ok := ctx.Value(UserRolesKey).([]string)
	return roles, ok
}

// GetUserPermissions extracts the user permissions from the request context.
func GetUserPermissions(ctx context.Context) ([]string, bool) {
	permissions, ok := ctx.Value(UserPermissionsKey).([]string)
	return permissions, ok
}

// GetAccountType extracts the account type from the request context.
func GetAccountType(ctx context.Context) (string, bool) {
	accountType, ok := ctx.Value(AccountTypeKey).(string)
	return accountType, ok
}
