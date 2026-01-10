/**
 * Verification Store
 * Manages verification requests state
 */

import { create } from 'zustand';
import { api, type Verification } from '../lib/api';

interface VerificationState {
  verifications: Verification[];
  isLoading: boolean;
  error: string | null;
  lastRefresh: number;

  // Actions
  fetchPendingVerifications: () => Promise<void>;
  refreshVerifications: () => Promise<void>;
  clearError: () => void;
}

// Auto-refresh interval: 30 seconds
const REFRESH_INTERVAL = 30 * 1000;

export const useVerificationStore = create<VerificationState>((set, get) => ({
  verifications: [],
  isLoading: false,
  error: null,
  lastRefresh: 0,

  fetchPendingVerifications: async () => {
    try {
      set({ isLoading: true, error: null });
      const verifications = await api.getPendingVerifications();
      set({
        verifications,
        isLoading: false,
        lastRefresh: Date.now(),
      });
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Failed to load verifications';
      set({
        error: errorMessage,
        isLoading: false,
      });
    }
  },

  refreshVerifications: async () => {
    const { lastRefresh, isLoading } = get();
    const now = Date.now();

    // Prevent refresh if already loading or refreshed recently
    if (isLoading || now - lastRefresh < 5000) {
      return;
    }

    await get().fetchPendingVerifications();
  },

  clearError: () => {
    set({ error: null });
  },
}));

// Auto-refresh verifications when authenticated
if (typeof window !== 'undefined') {
  setInterval(() => {
    const store = useVerificationStore.getState();
    // Only refresh if we have verifications loaded (user is authenticated)
    if (store.lastRefresh > 0 && !store.isLoading) {
      store.refreshVerifications();
    }
  }, REFRESH_INTERVAL);
}
