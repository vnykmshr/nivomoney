package handler

import (
	"net/http"

	"github.com/vnykmshr/nivo/shared/middleware"
)

// SetupRoutes configures all routes for the RBAC service using Go 1.22+ stdlib router.
func SetupRoutes(rbacHandler *RBACHandler, jwtSecret string) http.Handler {
	mux := http.NewServeMux()

	// Health check endpoint (public)
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","service":"rbac"}`))
	})

	// Setup auth middleware
	authConfig := middleware.AuthConfig{
		JWTSecret: jwtSecret,
		SkipPaths: []string{"/health"},
	}
	authMiddleware := middleware.Auth(authConfig)

	// Permission middleware for different access levels
	adminPerm := middleware.RequireAnyRole("admin", "super_admin")

	// ========================================================================
	// Role Management Endpoints (Admin Only)
	// ========================================================================

	mux.Handle("POST /api/v1/roles", authMiddleware(adminPerm(http.HandlerFunc(rbacHandler.CreateRole))))
	mux.Handle("GET /api/v1/roles", authMiddleware(http.HandlerFunc(rbacHandler.ListRoles)))
	mux.Handle("GET /api/v1/roles/{id}", authMiddleware(http.HandlerFunc(rbacHandler.GetRole)))
	mux.Handle("PUT /api/v1/roles/{id}", authMiddleware(adminPerm(http.HandlerFunc(rbacHandler.UpdateRole))))
	mux.Handle("DELETE /api/v1/roles/{id}", authMiddleware(adminPerm(http.HandlerFunc(rbacHandler.DeleteRole))))
	mux.Handle("GET /api/v1/roles/{id}/hierarchy", authMiddleware(http.HandlerFunc(rbacHandler.GetRoleHierarchy)))

	// ========================================================================
	// Permission Management Endpoints (Admin Only)
	// ========================================================================

	mux.Handle("POST /api/v1/permissions", authMiddleware(adminPerm(http.HandlerFunc(rbacHandler.CreatePermission))))
	mux.Handle("GET /api/v1/permissions", authMiddleware(http.HandlerFunc(rbacHandler.ListPermissions)))
	mux.Handle("GET /api/v1/permissions/{id}", authMiddleware(http.HandlerFunc(rbacHandler.GetPermission)))

	// ========================================================================
	// Role-Permission Assignment Endpoints (Admin Only)
	// ========================================================================

	mux.Handle("POST /api/v1/roles/{id}/permissions", authMiddleware(adminPerm(http.HandlerFunc(rbacHandler.AssignPermissionToRole))))
	mux.Handle("GET /api/v1/roles/{id}/permissions", authMiddleware(http.HandlerFunc(rbacHandler.GetRolePermissions)))
	mux.Handle("DELETE /api/v1/roles/{roleId}/permissions/{permissionId}", authMiddleware(adminPerm(http.HandlerFunc(rbacHandler.RemovePermissionFromRole))))

	// ========================================================================
	// User-Role Assignment Endpoints (Admin Only)
	// ========================================================================

	mux.Handle("POST /api/v1/users/{userId}/roles", authMiddleware(adminPerm(http.HandlerFunc(rbacHandler.AssignRoleToUser))))
	mux.Handle("GET /api/v1/users/{userId}/roles", authMiddleware(http.HandlerFunc(rbacHandler.GetUserRoles)))
	mux.Handle("DELETE /api/v1/users/{userId}/roles/{roleId}", authMiddleware(adminPerm(http.HandlerFunc(rbacHandler.RemoveRoleFromUser))))

	// User Permissions (authenticated users, used by Identity service)
	mux.Handle("GET /api/v1/users/{userId}/permissions", authMiddleware(http.HandlerFunc(rbacHandler.GetUserPermissions)))

	// ========================================================================
	// Permission Check Endpoints (Authenticated - used by services)
	// ========================================================================

	mux.Handle("POST /api/v1/check-permission", authMiddleware(http.HandlerFunc(rbacHandler.CheckPermission)))
	mux.Handle("POST /api/v1/check-permissions", authMiddleware(http.HandlerFunc(rbacHandler.CheckPermissions)))

	// Apply CORS middleware
	corsMiddleware := middleware.CORS(middleware.DefaultCORSConfig())
	return corsMiddleware(mux)
}
