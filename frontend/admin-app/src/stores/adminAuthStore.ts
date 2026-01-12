/**
 * Admin Authentication Store
 * Manages admin authentication state with stricter security controls
 */

import { create } from 'zustand';
import type { User } from '@nivo/shared';
import { adminApi } from '../lib/adminApi';
import { logAdminAction } from '../lib/auditLogger';

interface AdminAuthState {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  error: string | null;

  // Actions
  login: (identifier: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  loadUser: () => Promise<void>;
  clearError: () => void;

  // Session management
  lastActivityTime: number;
  updateActivity: () => void;
  checkSessionTimeout: () => boolean;
}

// Admin session timeout: 2 hours (stricter than user app)
const ADMIN_SESSION_TIMEOUT = 2 * 60 * 60 * 1000; // 2 hours in milliseconds

export const useAdminAuthStore = create<AdminAuthState>((set, get) => ({
  user: null,
  isAuthenticated: false,
  isLoading: false,
  error: null,
  lastActivityTime: Date.now(),

  login: async (identifier: string, password: string) => {
    try {
      set({ isLoading: true, error: null });

      const response = await adminApi.login({ identifier, password });

      // Verify user has admin permissions
      // TODO: Add proper role check when RBAC is fully implemented
      // For now, we trust the backend permission checks

      localStorage.setItem('admin_token', response.token);

      // Regenerate CSRF token on successful login
      adminApi.regenerateCSRF();

      set({
        user: response.user,
        isAuthenticated: true,
        isLoading: false,
        lastActivityTime: Date.now(),
      });

      // Audit log: Admin login
      logAdminAction.login(response.user.id, response.user.full_name);
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Login failed';
      set({
        error: errorMessage,
        isLoading: false,
        isAuthenticated: false,
        user: null,
      });
      throw error;
    }
  },

  logout: async () => {
    const { user } = get();
    try {
      // Audit log: Admin logout (before clearing state)
      if (user) {
        logAdminAction.logout(user.id, user.full_name);
      }

      await adminApi.logout();
    } catch {
      // Logout errors can be silently ignored - we're clearing local state anyway
    } finally {
      // Clear CSRF token on logout
      adminApi.clearCSRF();

      localStorage.removeItem('admin_token');
      set({
        user: null,
        isAuthenticated: false,
        error: null,
      });
    }
  },

  loadUser: async () => {
    const token = localStorage.getItem('admin_token');
    if (!token) {
      set({ isAuthenticated: false, user: null });
      return;
    }

    // Check session timeout
    const isExpired = get().checkSessionTimeout();
    if (isExpired) {
      await get().logout();
      set({ error: 'Session expired. Please login again.' });
      return;
    }

    try {
      set({ isLoading: true });
      const user = await adminApi.getProfile();

      set({
        user,
        isAuthenticated: true,
        isLoading: false,
        lastActivityTime: Date.now(),
      });
    } catch {
      // Token invalid or expired - clear it
      localStorage.removeItem('admin_token');
      set({
        user: null,
        isAuthenticated: false,
        isLoading: false,
      });
    }
  },

  updateActivity: () => {
    set({ lastActivityTime: Date.now() });
  },

  checkSessionTimeout: () => {
    const { lastActivityTime } = get();
    const now = Date.now();
    const timeSinceLastActivity = now - lastActivityTime;

    return timeSinceLastActivity > ADMIN_SESSION_TIMEOUT;
  },

  clearError: () => {
    set({ error: null });
  },
}));

// Set up activity tracking
if (typeof window !== 'undefined') {
  // Update activity on user interaction
  const updateActivity = () => {
    const store = useAdminAuthStore.getState();
    if (store.isAuthenticated) {
      store.updateActivity();
    }
  };

  window.addEventListener('mousemove', updateActivity);
  window.addEventListener('keypress', updateActivity);
  window.addEventListener('click', updateActivity);
  window.addEventListener('scroll', updateActivity);

  // Check session timeout every minute
  setInterval(() => {
    const store = useAdminAuthStore.getState();
    if (store.isAuthenticated && store.checkSessionTimeout()) {
      store.logout();
      alert('Session expired due to inactivity. Please login again.');
      window.location.href = '/login';
    }
  }, 60000); // Check every minute
}
