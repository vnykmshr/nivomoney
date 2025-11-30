import { useEffect, useCallback, useState, useMemo } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuthStore } from '../stores/authStore';
import { useWalletStore } from '../stores/walletStore';
import { useSSE } from '../hooks/useSSE';
import { WalletCard } from '../components/WalletCard';
import { TransactionList } from '../components/TransactionList';
import { TransactionFilters, type TransactionFilterValues } from '../components/TransactionFilters';
import { api } from '../lib/api';
import { formatCurrency } from '../lib/utils';
import type { Transaction, Wallet, KYCInfo } from '../types';

export function Dashboard() {
  const navigate = useNavigate();
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
  const [createError, setCreateError] = useState<string | null>(null);
  const [kycInfo, setKycInfo] = useState<KYCInfo | null>(null);
  const [kycLoading, setKycLoading] = useState(true);

  // Failsafe: Force kycLoading to false after 5 seconds to prevent stuck loading states
  useEffect(() => {
    const timeout = setTimeout(() => {
      setKycLoading(false);
    }, 5000);
    return () => clearTimeout(timeout);
  }, []);

  // Transaction filters
  const [filters, setFilters] = useState<TransactionFilterValues>({
    search: '',
    type: '',
    status: '',
    dateFrom: '',
    dateTo: '',
    amountMin: '',
    amountMax: '',
  });

  // Filter transactions
  const filteredTransactions = useMemo(() => {
    if (!transactions) return [];

    return transactions.filter(tx => {
      // Search filter
      if (filters.search) {
        const searchLower = filters.search.toLowerCase();
        const matchesDescription = tx.description?.toLowerCase().includes(searchLower);
        const matchesReference = tx.reference?.toLowerCase().includes(searchLower);
        const matchesId = tx.id.toLowerCase().includes(searchLower);
        if (!matchesDescription && !matchesReference && !matchesId) return false;
      }

      // Type filter
      if (filters.type && tx.type !== filters.type) return false;

      // Status filter
      if (filters.status && tx.status !== filters.status) return false;

      // Date range filter
      if (filters.dateFrom) {
        const txDate = new Date(tx.created_at).toISOString().split('T')[0];
        if (txDate < filters.dateFrom) return false;
      }
      if (filters.dateTo) {
        const txDate = new Date(tx.created_at).toISOString().split('T')[0];
        if (txDate > filters.dateTo) return false;
      }

      // Amount filter (convert to paise for comparison)
      const amountInRupees = tx.amount / 100;
      if (filters.amountMin && amountInRupees < parseFloat(filters.amountMin)) return false;
      if (filters.amountMax && amountInRupees > parseFloat(filters.amountMax)) return false;

      return true;
    });
  }, [transactions, filters]);

  const handleResetFilters = () => {
    setFilters({
      search: '',
      type: '',
      status: '',
      dateFrom: '',
      dateTo: '',
      amountMin: '',
      amountMax: '',
    });
  };

  // Fetch wallets and KYC status together to avoid race conditions
  useEffect(() => {
    const initializeDashboard = async () => {
      try {
        // Fetch both in parallel with explicit timeout handling
        await Promise.all([
          fetchWallets().catch(err => {
            console.error('Failed to fetch wallets:', err);
            return Promise.resolve(); // Don't block on wallet fetch failure
          }),
          api.getKYC()
            .then(kyc => setKycInfo(kyc))
            .catch(err => {
              // KYC might not exist yet - that's okay
              console.log('No KYC info found:', err);
              return Promise.resolve(); // Don't block on KYC fetch failure
            }),
        ]);
      } catch (err) {
        console.error('Dashboard initialization error:', err);
      } finally {
        // Always set kycLoading to false to prevent stuck loading state
        setKycLoading(false);
      }
    };

    initializeDashboard();
  }, [fetchWallets]);

  // Auto-select first wallet when wallets load
  useEffect(() => {
    if (!selectedWallet && wallets.length > 0) {
      // Prioritize default wallet, otherwise select first active wallet
      const defaultWallet = wallets.find(w => w.type === 'default' && w.status === 'active');
      const firstActiveWallet = wallets.find(w => w.status === 'active');
      const walletToSelect = defaultWallet || firstActiveWallet || wallets[0];
      if (walletToSelect) {
        selectWallet(walletToSelect.id);
      }
    }
  }, [wallets, selectedWallet, selectWallet]);

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
        type: 'default',
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

  // Render KYC status banner
  const renderKYCBanner = () => {
    if (kycLoading) return null;

    // No KYC submitted yet
    if (!kycInfo) {
      return (
        <div className="bg-yellow-50 border-l-4 border-yellow-400 p-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center">
              <div className="flex-shrink-0">
                <svg className="h-5 w-5 text-yellow-400" viewBox="0 0 20 20" fill="currentColor">
                  <path fillRule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
                </svg>
              </div>
              <div className="ml-3">
                <h3 className="text-sm font-medium text-yellow-800">Complete KYC Verification</h3>
                <p className="text-sm text-yellow-700 mt-1">
                  Submit your KYC documents to activate full account features and start transacting.
                </p>
              </div>
            </div>
            <button
              onClick={() => navigate('/kyc')}
              className="ml-4 flex-shrink-0 btn-primary"
            >
              Submit KYC
            </button>
          </div>
        </div>
      );
    }

    // KYC pending review
    if (kycInfo.status === 'pending') {
      return (
        <div className="bg-blue-50 border-l-4 border-blue-400 p-4">
          <div className="flex items-center">
            <div className="flex-shrink-0">
              <svg className="h-5 w-5 text-blue-400" viewBox="0 0 20 20" fill="currentColor">
                <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clipRule="evenodd" />
              </svg>
            </div>
            <div className="ml-3">
              <h3 className="text-sm font-medium text-blue-800">KYC Under Review</h3>
              <p className="text-sm text-blue-700 mt-1">
                Your KYC documents are being verified. We'll notify you once the review is complete.
              </p>
            </div>
          </div>
        </div>
      );
    }

    // KYC verified - show success badge
    if (kycInfo.status === 'verified') {
      return (
        <div className="bg-green-50 border-l-4 border-green-400 p-4">
          <div className="flex items-center">
            <div className="flex-shrink-0">
              <svg className="h-5 w-5 text-green-400" viewBox="0 0 20 20" fill="currentColor">
                <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
              </svg>
            </div>
            <div className="ml-3">
              <h3 className="text-sm font-medium text-green-800">KYC Verified</h3>
              <p className="text-sm text-green-700 mt-1">
                Your account is fully verified. You can now access all features.
              </p>
            </div>
          </div>
        </div>
      );
    }

    // KYC rejected - show error with resubmit option
    if (kycInfo.status === 'rejected') {
      return (
        <div className="bg-red-50 border-l-4 border-red-400 p-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center">
              <div className="flex-shrink-0">
                <svg className="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor">
                  <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
                </svg>
              </div>
              <div className="ml-3">
                <h3 className="text-sm font-medium text-red-800">KYC Rejected</h3>
                <p className="text-sm text-red-700 mt-1">
                  {kycInfo.rejection_reason || 'Your KYC documents were rejected. Please resubmit with correct information.'}
                </p>
              </div>
            </div>
            <button
              onClick={() => navigate('/kyc')}
              className="ml-4 flex-shrink-0 btn-primary bg-red-600 hover:bg-red-700"
            >
              Resubmit KYC
            </button>
          </div>
        </div>
      );
    }

    return null;
  };

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Navigation */}
      <nav className="bg-white shadow-sm">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between h-16 items-center">
            <h1 className="text-xl font-bold text-primary-600">Nivo Money</h1>
            <div className="flex items-center space-x-4">
              <span className="text-gray-700">{user?.full_name}</span>
              <button onClick={() => navigate('/profile')} className="btn-secondary">
                Profile
              </button>
              <button onClick={logout} className="btn-secondary">
                Logout
              </button>
            </div>
          </div>
        </div>
      </nav>

      {/* KYC Status Banner */}
      {renderKYCBanner()}

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Welcome Section with User Info */}
        <div className="mb-8">
          <div className="flex justify-between items-start">
            <div>
              <h2 className="text-3xl font-bold text-gray-900">Welcome back, {user?.full_name}!</h2>
              <p className="text-gray-600 mt-1">Here's your financial overview</p>
            </div>
            <div className="text-right">
              <div className="text-sm text-gray-600">Phone</div>
              <div className="font-medium text-gray-900">{user?.phone}</div>
              {user?.email && (
                <>
                  <div className="text-sm text-gray-600 mt-2">Email</div>
                  <div className="font-medium text-gray-900">{user?.email}</div>
                </>
              )}
            </div>
          </div>
        </div>

        {/* Loading State - show loading until both wallet and KYC data are ready */}
        {(isLoading || kycLoading) && wallets.length === 0 && (
          <div className="card text-center py-12">
            <div className="text-gray-500">Loading your dashboard...</div>
          </div>
        )}

        {/* Wallets Section */}
        {!isLoading && !kycLoading && wallets.length === 0 && (
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
                <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
                  <h4 className="text-sm font-medium text-blue-900 mb-2">Default Wallet</h4>
                  <p className="text-sm text-blue-800">
                    A default INR wallet will be created for your account. You can use it for all transactions including deposits, withdrawals, and transfers.
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
            {/* Summary Cards */}
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
              <div className="card bg-gradient-to-br from-primary-500 to-primary-600 text-white">
                <div className="text-sm opacity-90 mb-2">Total Balance</div>
                <div className="text-3xl font-bold">
                  {formatCurrency(wallets.reduce((sum, w) => sum + w.balance, 0))}
                </div>
                <div className="text-sm opacity-75 mt-2">Across {wallets.length} wallet{wallets.length !== 1 ? 's' : ''}</div>
              </div>

              <div className="card bg-gradient-to-br from-green-500 to-green-600 text-white">
                <div className="text-sm opacity-90 mb-2">Available Balance</div>
                <div className="text-3xl font-bold">
                  {formatCurrency(wallets.reduce((sum, w) => sum + w.available_balance, 0))}
                </div>
                <div className="text-sm opacity-75 mt-2">Ready to use</div>
              </div>

              <div className="card bg-gradient-to-br from-blue-500 to-blue-600 text-white">
                <div className="text-sm opacity-90 mb-2">Recent Transactions</div>
                <div className="text-3xl font-bold">{transactions.length}</div>
                <div className="text-sm opacity-75 mt-2">
                  {selectedWallet ? `For selected wallet` : 'Select a wallet to view'}
                </div>
              </div>
            </div>

            {/* Wallets Grid */}
            <div className="mb-8">
              <h3 className="text-xl font-semibold text-gray-900 mb-4">Your Wallets</h3>
              <p className="text-sm text-gray-600 mb-4">Click on a wallet to view its transactions below</p>
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                {wallets.map(wallet => (
                  <WalletCard
                    key={wallet.id}
                    wallet={wallet}
                    isSelected={selectedWallet?.id === wallet.id}
                    onClick={() => {
                      selectWallet(wallet.id);
                      // Smooth scroll to transactions section
                      setTimeout(() => {
                        document.getElementById('transactions-section')?.scrollIntoView({
                          behavior: 'smooth',
                          block: 'start'
                        });
                      }, 100);
                    }}
                  />
                ))}
              </div>
            </div>

            {/* Quick Actions */}
            <div className="mb-8">
              <h3 className="text-xl font-semibold text-gray-900 mb-4">Quick Actions</h3>
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
                <button
                  onClick={() => window.location.href = '/add-money'}
                  className="btn-primary py-4 text-lg"
                >
                  💰 Add Money
                </button>
                <button
                  onClick={() => window.location.href = '/send'}
                  className="btn-primary py-4 text-lg"
                >
                  💸 Send Money
                </button>
                <button
                  onClick={() => window.location.href = '/beneficiaries'}
                  className="btn-primary py-4 text-lg"
                >
                  👥 Saved Recipients
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
              <div id="transactions-section" className="mb-8">
                <h3 className="text-xl font-semibold text-gray-900 mb-4">
                  Transactions - {selectedWallet.type.charAt(0).toUpperCase() + selectedWallet.type.slice(1)} Wallet
                </h3>

                {/* Transaction Filters */}
                {transactions.length > 0 && (
                  <TransactionFilters
                    filters={filters}
                    onFilterChange={setFilters}
                    onReset={handleResetFilters}
                  />
                )}

                {isLoading ? (
                  <div className="card text-center py-8 text-gray-500">
                    Loading transactions...
                  </div>
                ) : (
                  <>
                    <TransactionList transactions={filteredTransactions} walletId={selectedWallet.id} />
                    {transactions.length > 0 && filteredTransactions.length === 0 && (
                      <div className="card text-center py-8">
                        <p className="text-gray-500">No transactions match your filters.</p>
                        <button
                          onClick={handleResetFilters}
                          className="btn-secondary mt-4"
                        >
                          Clear Filters
                        </button>
                      </div>
                    )}
                  </>
                )}
              </div>
            )}
          </>
        )}
      </main>
    </div>
  );
}
