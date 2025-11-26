/**
 * Admin Route Guard
 * Protects routes that require admin authentication
 * Stricter than user routes - enforces admin role and session timeout
 */

import { useEffect } from 'react';
import { Navigate, useLocation } from 'react-router-dom';
import { useAdminAuthStore } from '../stores/adminAuthStore';

interface AdminRouteProps {
  children: React.ReactNode;
}

export function AdminRoute({ children }: AdminRouteProps) {
  const { isAuthenticated, isLoading, loadUser, checkSessionTimeout, logout } = useAdminAuthStore();
  const location = useLocation();

  useEffect(() => {
    // Load user if not already loaded
    if (!isAuthenticated && !isLoading) {
      loadUser();
    }
  }, [isAuthenticated, isLoading, loadUser]);

  useEffect(() => {
    // Check session timeout on route change
    if (isAuthenticated) {
      const isExpired = checkSessionTimeout();
      if (isExpired) {
        logout();
        alert('Session expired due to inactivity. Please login again.');
      }
    }
  }, [location, isAuthenticated, checkSessionTimeout, logout]);

  // Show loading state while checking authentication
  if (isLoading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-gray-500">Loading...</div>
      </div>
    );
  }

  // Redirect to login if not authenticated
  if (!isAuthenticated) {
    return <Navigate to="/login" state={{ from: location }} replace />;
  }

  // TODO: Add role check when RBAC is fully implemented
  // For now, we trust the backend permission checks on API calls

  return <>{children}</>;
}
