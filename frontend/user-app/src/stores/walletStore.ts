import { create } from 'zustand';
import type { Wallet, Transaction } from '../types';
import { api } from '../lib/api';

interface WalletState {
  wallets: Wallet[];
  selectedWallet: Wallet | null;
  transactions: Transaction[];
  isLoading: boolean;
  error: string | null;

  fetchWallets: () => Promise<void>;
  selectWallet: (walletId: string) => void;
  fetchTransactions: (walletId: string) => Promise<void>;
  setError: (error: string | null) => void;

  // SSE event handlers
  updateWalletFromEvent: (walletData: Partial<Wallet> & { id: string }) => void;
  addTransactionFromEvent: (transaction: Transaction) => void;
  updateTransactionFromEvent: (transactionData: Partial<Transaction> & { id: string }) => void;
}

export const useWalletStore = create<WalletState>(set => ({
  wallets: [],
  selectedWallet: null,
  transactions: [],
  isLoading: false,
  error: null,

  fetchWallets: async () => {
    set({ isLoading: true, error: null });
    try {
      const wallets = await api.getWallets();
      // Pre-select default active wallet, or first active wallet, or first wallet
      const defaultWallet = wallets.find(w => w.type === 'default' && w.status === 'active');
      const firstActiveWallet = wallets.find(w => w.status === 'active');
      const selectedWallet = defaultWallet || firstActiveWallet || (wallets.length > 0 ? wallets[0] : null);

      set({
        wallets,
        selectedWallet,
        isLoading: false,
      });
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Failed to fetch wallets';
      set({ error: errorMessage, isLoading: false });
      throw error;
    }
  },

  selectWallet: (walletId: string) => {
    set(state => ({
      selectedWallet: state.wallets.find(w => w.id === walletId) || null,
    }));
  },

  fetchTransactions: async (walletId: string) => {
    set({ isLoading: true, error: null });
    try {
      const transactions = await api.getTransactions(walletId);
      set({ transactions, isLoading: false });
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Failed to fetch transactions';
      set({ error: errorMessage, isLoading: false });
      throw error;
    }
  },

  setError: (error: string | null) => {
    set({ error });
  },

  // SSE event handlers
  updateWalletFromEvent: (walletData: Partial<Wallet> & { id: string }) => {
    set(state => {
      const updatedWallets = state.wallets.map(w =>
        w.id === walletData.id ? { ...w, ...walletData } : w
      );
      const updatedSelectedWallet =
        state.selectedWallet?.id === walletData.id
          ? { ...state.selectedWallet, ...walletData }
          : state.selectedWallet;
      return {
        wallets: updatedWallets,
        selectedWallet: updatedSelectedWallet,
      };
    });
  },

  addTransactionFromEvent: (transaction: Transaction) => {
    set(state => {
      // Check if transaction already exists
      const exists = state.transactions.some(t => t.id === transaction.id);
      if (exists) {
        // Update existing transaction
        return {
          transactions: state.transactions.map(t =>
            t.id === transaction.id ? transaction : t
          ),
        };
      }
      // Add new transaction at the beginning
      return {
        transactions: [transaction, ...state.transactions],
      };
    });
  },

  updateTransactionFromEvent: (transactionData: Partial<Transaction> & { id: string }) => {
    set(state => ({
      transactions: state.transactions.map(t =>
        t.id === transactionData.id ? { ...t, ...transactionData } : t
      ),
    }));
  },
}));
