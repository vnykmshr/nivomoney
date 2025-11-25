import { useEffect, useCallback, useState } from 'react';
import { useAuthStore } from '../stores/authStore';
import { useWalletStore } from '../stores/walletStore';
import { useSSE } from '../hooks/useSSE';
import { WalletCard } from '../components/WalletCard';
import { TransactionList } from '../components/TransactionList';
import { api } from '../lib/api';
import type { Transaction, Wallet } from '../types';

export function Dashboard() {
  const { user, logout } = useAuthStore();
  const {
    wallets,
    selectedWallet,
    transactions,
    isLoading,
    fetchWallets,
    fetchTransactions,
    selectWallet,
    updateWalletFromEvent,
    addTransactionFromEvent,
    updateTransactionFromEvent,
  } = useWalletStore();

  const [isCreatingWallet, setIsCreatingWallet] = useState(false);
  const [walletType, setWalletType] = useState<'savings' | 'current' | 'investment'>('savings');
  const [createError, setCreateError] = useState<string | null>(null);

  useEffect(() => {
    fetchWallets().catch(err => console.error('Failed to fetch wallets:', err));
  }, [fetchWallets]);

  useEffect(() => {
    if (selectedWallet) {
      fetchTransactions(selectedWallet.id).catch(err =>
        console.error('Failed to fetch transactions:', err)
      );
    }
  }, [selectedWallet, fetchTransactions]);

  // Handle wallet creation
  const handleCreateWallet = async () => {
    if (!user?.id) return;

    setIsCreatingWallet(true);
    setCreateError(null);

    try {
      await api.createWallet({
        user_id: user.id,
        type: walletType,
        currency: 'INR',
      });

      // Refetch wallets to show the new wallet
      await fetchWallets();
    } catch (err) {
      setCreateError(err instanceof Error ? err.message : 'Failed to create wallet');
    } finally {
      setIsCreatingWallet(false);
    }
  };

  // Handle SSE events
  const handleSSEEvent = useCallback(
    (event: { topic: string; event_type: string; data: Record<string, unknown> }) => {
      console.log('Received SSE event:', event);

      if (event.topic === 'wallets') {
        if (event.event_type === 'wallet.updated' || event.event_type === 'wallet.created') {
          updateWalletFromEvent(event.data as unknown as Partial<Wallet> & { id: string });
        }
      } else if (event.topic === 'transactions') {
        if (event.event_type === 'transaction.created') {
          addTransactionFromEvent(event.data as unknown as Transaction);
          // Also refetch wallets to update balance
          fetchWallets().catch(err => console.error('Failed to refetch wallets:', err));
        } else if (event.event_type === 'transaction.updated') {
          updateTransactionFromEvent(event.data as unknown as Partial<Transaction> & { id: string });
          // Also refetch wallets to update balance
          fetchWallets().catch(err => console.error('Failed to refetch wallets:', err));
        }
      }
    },
    [updateWalletFromEvent, addTransactionFromEvent, updateTransactionFromEvent, fetchWallets]
  );

  // Connect to SSE for real-time updates
  useSSE({
    topics: ['wallets', 'transactions'],
    onEvent: handleSSEEvent,
    onError: error => console.error('SSE error:', error),
    enabled: !!user,
  });

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Navigation */}
      <nav className="bg-white shadow-sm">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between h-16 items-center">
            <h1 className="text-xl font-bold text-primary-600">Nivo Money</h1>
            <div className="flex items-center space-x-4">
              <span className="text-gray-700">{user?.full_name}</span>
              <button onClick={logout} className="btn-secondary">
                Logout
              </button>
            </div>
          </div>
        </div>
      </nav>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Welcome Section */}
        <div className="mb-8">
          <h2 className="text-3xl font-bold text-gray-900">Welcome back, {user?.full_name}!</h2>
          <p className="text-gray-600 mt-1">Here's your financial overview</p>
        </div>

        {/* Loading State */}
        {isLoading && wallets.length === 0 && (
          <div className="card text-center py-12">
            <div className="text-gray-500">Loading your wallets...</div>
          </div>
        )}

        {/* Wallets Section */}
        {!isLoading && wallets.length === 0 && (
          <div className="card py-12">
            <div className="max-w-md mx-auto">
              <h3 className="text-2xl font-bold text-gray-900 mb-2 text-center">Create Your First Wallet</h3>
              <p className="text-gray-600 mb-6 text-center">Start managing your money with a new wallet</p>

              {createError && (
                <div className="mb-4 p-4 bg-red-100 text-red-800 rounded-lg text-sm">
                  {createError}
                </div>
              )}

              <div className="space-y-4">
                <div>
                  <label htmlFor="walletType" className="block text-sm font-medium text-gray-700 mb-2">
                    Wallet Type
                  </label>
                  <select
                    id="walletType"
                    value={walletType}
                    onChange={e => setWalletType(e.target.value as 'savings' | 'current' | 'investment')}
                    className="input-field"
                    disabled={isCreatingWallet}
                  >
                    <option value="savings">Savings Wallet</option>
                    <option value="current">Current Wallet</option>
                    <option value="investment">Investment Wallet</option>
                  </select>
                  <p className="text-sm text-gray-500 mt-1">
                    {walletType === 'savings' && 'For personal savings and everyday transactions'}
                    {walletType === 'current' && 'For business and frequent transactions'}
                    {walletType === 'investment' && 'For investment and long-term growth'}
                  </p>
                </div>

                <button
                  onClick={handleCreateWallet}
                  className="btn-primary w-full"
                  disabled={isCreatingWallet}
                >
                  {isCreatingWallet ? 'Creating Wallet...' : 'Create Wallet'}
                </button>
              </div>
            </div>
          </div>
        )}

        {wallets.length > 0 && (
          <>
            {/* Wallets Grid */}
            <div className="mb-8">
              <h3 className="text-xl font-semibold text-gray-900 mb-4">Your Wallets</h3>
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                {wallets.map(wallet => (
                  <WalletCard
                    key={wallet.id}
                    wallet={wallet}
                    isSelected={selectedWallet?.id === wallet.id}
                    onClick={() => selectWallet(wallet.id)}
                  />
                ))}
              </div>
            </div>

            {/* Quick Actions */}
            <div className="mb-8">
              <h3 className="text-xl font-semibold text-gray-900 mb-4">Quick Actions</h3>
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                <button
                  onClick={() => window.location.href = '/send'}
                  className="btn-primary py-4 text-lg"
                >
                  💸 Send Money
                </button>
                <button
                  onClick={() => window.location.href = '/deposit'}
                  className="btn-primary py-4 text-lg"
                >
                  💰 Deposit
                </button>
                <button
                  onClick={() => window.location.href = '/withdraw'}
                  className="btn-primary py-4 text-lg"
                >
                  🏧 Withdraw
                </button>
              </div>
            </div>

            {/* Transactions Section */}
            {selectedWallet && (
              <div className="mb-8">
                <h3 className="text-xl font-semibold text-gray-900 mb-4">
                  Transactions - {selectedWallet.type.charAt(0).toUpperCase() + selectedWallet.type.slice(1)} Wallet
                </h3>
                {isLoading ? (
                  <div className="card text-center py-8 text-gray-500">
                    Loading transactions...
                  </div>
                ) : (
                  <TransactionList transactions={transactions} walletId={selectedWallet.id} />
                )}
              </div>
            )}
          </>
        )}
      </main>
    </div>
  );
}
