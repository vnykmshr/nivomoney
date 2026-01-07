import { useEffect, useCallback, useState, useMemo } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuthStore } from '../stores/authStore';
import { useWalletStore } from '../stores/walletStore';
import { useSSE } from '../hooks/useSSE';
import {
  AppLayout,
  BalanceCard,
  QuickActionGrid,
  QuickActionIcons,
  TransactionList,
  TransactionFilters,
  type TransactionFilterValues,
  type QuickAction,
} from '../components';
import { Alert, Card, CardTitle, Button, Skeleton } from '../../../shared/components';
import { api } from '../lib/api';
import type { Transaction, Wallet, KYCInfo } from '../types';

export function Dashboard() {
  const navigate = useNavigate();
  const { user } = useAuthStore();
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

  // Failsafe timeout for KYC loading
  useEffect(() => {
    const timeout = setTimeout(() => setKycLoading(false), 5000);
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
      if (filters.search) {
        const searchLower = filters.search.toLowerCase();
        const matchesDescription = tx.description?.toLowerCase().includes(searchLower);
        const matchesReference = tx.reference?.toLowerCase().includes(searchLower);
        const matchesId = tx.id.toLowerCase().includes(searchLower);
        if (!matchesDescription && !matchesReference && !matchesId) return false;
      }
      if (filters.type && tx.type !== filters.type) return false;
      if (filters.status && tx.status !== filters.status) return false;
      if (filters.dateFrom) {
        const txDate = new Date(tx.created_at).toISOString().split('T')[0];
        if (txDate < filters.dateFrom) return false;
      }
      if (filters.dateTo) {
        const txDate = new Date(tx.created_at).toISOString().split('T')[0];
        if (txDate > filters.dateTo) return false;
      }
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

  // Initialize dashboard data
  useEffect(() => {
    const initializeDashboard = async () => {
      try {
        await Promise.all([
          fetchWallets().catch(err => console.error('Failed to fetch wallets:', err)),
          api.getKYC()
            .then(kyc => setKycInfo(kyc))
            .catch(() => {}),
        ]);
      } catch (err) {
        console.error('Dashboard initialization error:', err);
      } finally {
        setKycLoading(false);
      }
    };
    initializeDashboard();
  }, [fetchWallets]);

  // Auto-select wallet
  useEffect(() => {
    if (!selectedWallet && wallets.length > 0) {
      const defaultWallet = wallets.find(w => w.type === 'default' && w.status === 'active');
      const firstActiveWallet = wallets.find(w => w.status === 'active');
      const walletToSelect = defaultWallet || firstActiveWallet || wallets[0];
      if (walletToSelect) selectWallet(walletToSelect.id);
    }
  }, [wallets, selectedWallet, selectWallet]);

  // Fetch transactions for selected wallet
  useEffect(() => {
    if (selectedWallet) {
      fetchTransactions(selectedWallet.id).catch(err =>
        console.error('Failed to fetch transactions:', err)
      );
    }
  }, [selectedWallet, fetchTransactions]);

  // Create wallet handler
  const handleCreateWallet = async () => {
    if (!user?.id) return;
    setIsCreatingWallet(true);
    setCreateError(null);
    try {
      await api.createWallet({ user_id: user.id, type: 'default', currency: 'INR' });
      await fetchWallets();
    } catch (err) {
      setCreateError(err instanceof Error ? err.message : 'Failed to create wallet');
    } finally {
      setIsCreatingWallet(false);
    }
  };

  // SSE event handler
  const handleSSEEvent = useCallback(
    (event: { topic: string; event_type: string; data: Record<string, unknown> }) => {
      if (event.topic === 'wallets') {
        if (event.event_type === 'wallet.updated' || event.event_type === 'wallet.created') {
          updateWalletFromEvent(event.data as unknown as Partial<Wallet> & { id: string });
        }
      } else if (event.topic === 'transactions') {
        if (event.event_type === 'transaction.created') {
          addTransactionFromEvent(event.data as unknown as Transaction);
          fetchWallets().catch(err => console.error('Failed to refetch wallets:', err));
        } else if (event.event_type === 'transaction.updated') {
          updateTransactionFromEvent(event.data as unknown as Partial<Transaction> & { id: string });
          fetchWallets().catch(err => console.error('Failed to refetch wallets:', err));
        }
      }
    },
    [updateWalletFromEvent, addTransactionFromEvent, updateTransactionFromEvent, fetchWallets]
  );

  // SSE connection
  useSSE({
    topics: ['wallets', 'transactions'],
    onEvent: handleSSEEvent,
    onError: error => console.error('SSE error:', error),
    enabled: !!user,
  });

  // Quick actions
  const quickActions: QuickAction[] = [
    { id: 'add', label: 'Add Money', href: '/add-money', icon: QuickActionIcons.addMoney, color: 'success' },
    { id: 'send', label: 'Send Money', href: '/send', icon: QuickActionIcons.sendMoney, color: 'primary' },
    { id: 'withdraw', label: 'Withdraw', href: '/withdraw', icon: QuickActionIcons.withdraw, color: 'warning' },
    { id: 'beneficiaries', label: 'Recipients', href: '/beneficiaries', icon: QuickActionIcons.beneficiaries, color: 'neutral' },
  ];

  // KYC banner
  const renderKYCBanner = () => {
    if (kycLoading) return null;

    if (!kycInfo) {
      return (
        <Alert variant="warning" title="Complete KYC Verification">
          <div className="flex items-center justify-between">
            <p>Submit your KYC documents to activate full account features.</p>
            <Button size="sm" onClick={() => navigate('/kyc')} className="ml-4 flex-shrink-0">
              Submit KYC
            </Button>
          </div>
        </Alert>
      );
    }

    if (kycInfo.status === 'pending') {
      return (
        <Alert variant="info" title="KYC Under Review">
          Your KYC documents are being verified. We'll notify you once complete.
        </Alert>
      );
    }

    if (kycInfo.status === 'verified') {
      return (
        <Alert variant="success" title="KYC Verified">
          Your account is fully verified. You can access all features.
        </Alert>
      );
    }

    if (kycInfo.status === 'rejected') {
      return (
        <Alert variant="error" title="KYC Rejected">
          <div className="flex items-center justify-between">
            <p>{kycInfo.rejection_reason || 'Your documents were rejected. Please resubmit.'}</p>
            <Button size="sm" variant="danger" onClick={() => navigate('/kyc')} className="ml-4 flex-shrink-0">
              Resubmit
            </Button>
          </div>
        </Alert>
      );
    }

    return null;
  };

  // Calculate totals
  const totalBalance = wallets.reduce((sum, w) => sum + w.balance, 0);
  const availableBalance = wallets.reduce((sum, w) => sum + w.available_balance, 0);

  return (
    <AppLayout>
      <div className="max-w-5xl mx-auto px-4 py-6 space-y-6">
        {/* Welcome + KYC Banner */}
        <div className="space-y-4">
          <div>
            <h2 className="text-2xl font-bold text-[var(--text-primary)]">
              Welcome back, {user?.full_name?.split(' ')[0]}!
            </h2>
            <p className="text-[var(--text-secondary)]">
              Here's your financial overview
            </p>
          </div>
          {renderKYCBanner()}
        </div>

        {/* Loading State */}
        {(isLoading || kycLoading) && wallets.length === 0 && (
          <Card>
            <div className="space-y-4">
              <Skeleton className="h-8 w-48" />
              <Skeleton className="h-32" />
              <div className="grid grid-cols-4 gap-3">
                <Skeleton className="h-24" />
                <Skeleton className="h-24" />
                <Skeleton className="h-24" />
                <Skeleton className="h-24" />
              </div>
            </div>
          </Card>
        )}

        {/* No Wallet State */}
        {!isLoading && !kycLoading && wallets.length === 0 && (
          <Card padding="lg">
            <div className="max-w-md mx-auto text-center">
              <div className="w-16 h-16 mx-auto mb-4 rounded-full bg-[var(--surface-brand-subtle)] flex items-center justify-center">
                <svg className="w-8 h-8 text-[var(--interactive-primary)]" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 10h18M7 15h1m4 0h1m-7 4h12a3 3 0 003-3V8a3 3 0 00-3-3H6a3 3 0 00-3 3v8a3 3 0 003 3z" />
                </svg>
              </div>
              <CardTitle className="mb-2">Create Your First Wallet</CardTitle>
              <p className="text-[var(--text-secondary)] mb-6">
                Start managing your money with a new wallet
              </p>

              {createError && (
                <Alert variant="error" className="mb-4">{createError}</Alert>
              )}

              <Button
                onClick={handleCreateWallet}
                loading={isCreatingWallet}
                size="lg"
                className="w-full"
              >
                Create Wallet
              </Button>
            </div>
          </Card>
        )}

        {/* Main Dashboard Content */}
        {wallets.length > 0 && selectedWallet && (
          <>
            {/* Balance Card */}
            <BalanceCard
              wallet={selectedWallet}
              totalBalance={totalBalance}
              availableBalance={availableBalance}
            />

            {/* Quick Actions */}
            <section>
              <h3 className="text-lg font-semibold text-[var(--text-primary)] mb-3">
                Quick Actions
              </h3>
              <QuickActionGrid actions={quickActions} />
            </section>

            {/* Recent Transactions */}
            <section id="transactions-section">
              <div className="flex items-center justify-between mb-4">
                <h3 className="text-lg font-semibold text-[var(--text-primary)]">
                  Recent Transactions
                </h3>
                {transactions.length > 0 && (
                  <span className="text-sm text-[var(--text-muted)]">
                    {filteredTransactions.length} of {transactions.length}
                  </span>
                )}
              </div>

              {transactions.length > 0 && (
                <TransactionFilters
                  filters={filters}
                  onFilterChange={setFilters}
                  onReset={handleResetFilters}
                />
              )}

              {isLoading ? (
                <Card>
                  <div className="space-y-3">
                    {[1, 2, 3].map(i => (
                      <Skeleton key={i} className="h-16" />
                    ))}
                  </div>
                </Card>
              ) : (
                <>
                  <TransactionList
                    transactions={filteredTransactions}
                    walletId={selectedWallet.id}
                  />
                  {transactions.length > 0 && filteredTransactions.length === 0 && (
                    <Card className="text-center py-8">
                      <p className="text-[var(--text-muted)] mb-4">
                        No transactions match your filters.
                      </p>
                      <Button variant="secondary" onClick={handleResetFilters}>
                        Clear Filters
                      </Button>
                    </Card>
                  )}
                </>
              )}
            </section>
          </>
        )}
      </div>
    </AppLayout>
  );
}
