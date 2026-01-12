/**
 * User-Admin Authentication Store
 * Manages authentication state for User-Admin portal
 */

import { create } from 'zustand';
import type { User } from '@nivo/shared';
import { api, type PairedUser } from '../lib/api';

interface AuthState {
  user: User | null;
  pairedUser: PairedUser | null;
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

// User-Admin session timeout: 1 hour
const SESSION_TIMEOUT = 60 * 60 * 1000; // 1 hour in milliseconds

export const useAuthStore = create<AuthState>((set, get) => ({
  user: null,
  pairedUser: null,
  isAuthenticated: false,
  isLoading: false,
  error: null,
  lastActivityTime: Date.now(),

  login: async (identifier: string, password: string) => {
    try {
      set({ isLoading: true, error: null });

      const response = await api.login({ identifier, password });

      localStorage.setItem('user_admin_token', response.token);

      // Load paired user info (optional, may fail if not set up)
      let pairedUser: PairedUser | null = null;
      try {
        pairedUser = await api.getPairedUser();
      } catch {
        // Paired user not configured - this is optional
      }

      set({
        user: response.user,
        pairedUser,
        isAuthenticated: true,
        isLoading: false,
        lastActivityTime: Date.now(),
      });
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Login failed';
      set({
        error: errorMessage,
        isLoading: false,
        isAuthenticated: false,
        user: null,
        pairedUser: null,
      });
      throw error;
    }
  },

  logout: async () => {
    try {
      await api.logout();
    } catch {
      // Logout errors can be silently ignored - we're clearing local state anyway
    } finally {
      localStorage.removeItem('user_admin_token');
      set({
        user: null,
        pairedUser: null,
        isAuthenticated: false,
        error: null,
      });
    }
  },

  loadUser: async () => {
    const token = localStorage.getItem('user_admin_token');
    if (!token) {
      set({ isAuthenticated: false, user: null, pairedUser: null });
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
      const user = await api.getProfile();

      // Load paired user info (optional, may fail if not set up)
      let pairedUser: PairedUser | null = null;
      try {
        pairedUser = await api.getPairedUser();
      } catch {
        // Paired user not configured - this is optional
      }

      set({
        user,
        pairedUser,
        isAuthenticated: true,
        isLoading: false,
        lastActivityTime: Date.now(),
      });
    } catch {
      // Token invalid or expired - clear it
      localStorage.removeItem('user_admin_token');
      set({
        user: null,
        pairedUser: null,
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
    return timeSinceLastActivity > SESSION_TIMEOUT;
  },

  clearError: () => {
    set({ error: null });
  },
}));

// Set up activity tracking
if (typeof window !== 'undefined') {
  const updateActivity = () => {
    const store = useAuthStore.getState();
    if (store.isAuthenticated) {
      store.updateActivity();
    }
  };

  window.addEventListener('mousemove', updateActivity);
  window.addEventListener('keypress', updateActivity);
  window.addEventListener('click', updateActivity);

  // Check session timeout every minute
  setInterval(() => {
    const store = useAuthStore.getState();
    if (store.isAuthenticated && store.checkSessionTimeout()) {
      store.logout();
      alert('Session expired due to inactivity. Please login again.');
      window.location.href = '/login';
    }
  }, 60000);
}
