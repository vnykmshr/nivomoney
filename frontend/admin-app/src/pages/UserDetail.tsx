/**
 * User Detail Page
 * Comprehensive view of user profile, KYC, wallets, and transactions
 */

import { useState, useEffect, useCallback } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { AdminLayout, TransactionDetailModal } from '../components';
import { adminApi } from '../lib/adminApi';
import type { User, Wallet, Transaction } from '@nivo/shared';
import {
  Card,
  CardTitle,
  Button,
  Alert,
  Badge,
  Skeleton,
  FormField,
} from '../../../shared/components';
import {
  cn,
  getStatusVariant,
  getKYCStatusVariant,
  getWalletStatusVariant,
  getTransactionStatusVariant,
  getTransactionTypeVariant,
} from '../../../shared/lib';

type Tab = 'profile' | 'kyc' | 'wallets' | 'transactions';

export function UserDetail() {
  const { userId } = useParams<{ userId: string }>();
  const navigate = useNavigate();
  const [activeTab, setActiveTab] = useState<Tab>('profile');
  const [user, setUser] = useState<User | null>(null);
  const [wallets, setWallets] = useState<Wallet[]>([]);
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isLoadingTransactions, setIsLoadingTransactions] = useState(false);
  const [transactionsLoaded, setTransactionsLoaded] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [selectedTransactionId, setSelectedTransactionId] = useState<string | null>(null);

  // Wallet action modals
  const [showFreezeModal, setShowFreezeModal] = useState(false);
  const [showUnfreezeModal, setShowUnfreezeModal] = useState(false);
  const [showCloseModal, setShowCloseModal] = useState(false);
  const [selectedWallet, setSelectedWallet] = useState<Wallet | null>(null);
  const [actionReason, setActionReason] = useState('');
  const [isProcessing, setIsProcessing] = useState(false);

  // User suspension modals
  const [showSuspendModal, setShowSuspendModal] = useState(false);
  const [showUnsuspendModal, setShowUnsuspendModal] = useState(false);
  const [suspensionReason, setSuspensionReason] = useState('');

  useEffect(() => {
    if (!userId) {
      navigate('/');
      return;
    }
    loadUserData();
  }, [userId]);

  const loadUserData = async () => {
    if (!userId) return;

    try {
      setIsLoading(true);
      setError(null);

      // Load user details and wallets in parallel
      const [userData, walletsData] = await Promise.all([
        adminApi.getUserDetails(userId),
        adminApi.getUserWallets(userId),
      ]);

      setUser(userData);
      setWallets(walletsData);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load user data');
    } finally {
      setIsLoading(false);
    }
  };

  // Load transactions on demand when tab is selected
  const loadTransactions = useCallback(async () => {
    if (!userId || transactionsLoaded) return; // Already loaded

    try {
      setIsLoadingTransactions(true);
      const txData = await adminApi.searchTransactions({ user_id: userId, limit: 50 });
      setTransactions(txData);
      setTransactionsLoaded(true);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load transactions');
    } finally {
      setIsLoadingTransactions(false);
    }
  }, [userId, transactionsLoaded]);

  // Load transactions when tab is selected
  useEffect(() => {
    if (activeTab === 'transactions') {
      loadTransactions();
    }
  }, [activeTab, loadTransactions]);

  const handleFreezeWallet = (wallet: Wallet) => {
    setSelectedWallet(wallet);
    setActionReason('');
    setShowFreezeModal(true);
  };

  const handleUnfreezeWallet = (wallet: Wallet) => {
    setSelectedWallet(wallet);
    setShowUnfreezeModal(true);
  };

  const handleCloseWallet = (wallet: Wallet) => {
    setSelectedWallet(wallet);
    setActionReason('');
    setShowCloseModal(true);
  };

  const confirmFreezeWallet = async () => {
    if (!selectedWallet || !actionReason.trim()) {
      setError('Please provide a reason for freezing this wallet');
      return;
    }

    if (actionReason.length < 10) {
      setError('Reason must be at least 10 characters');
      return;
    }

    try {
      setIsProcessing(true);
      setError(null);

      await adminApi.freezeWallet(selectedWallet.id, { reason: actionReason });

      // Reload wallets
      const walletsData = await adminApi.getUserWallets(userId!);
      setWallets(walletsData);

      setShowFreezeModal(false);
      setSelectedWallet(null);
      setActionReason('');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to freeze wallet');
    } finally {
      setIsProcessing(false);
    }
  };

  const confirmUnfreezeWallet = async () => {
    if (!selectedWallet) return;

    try {
      setIsProcessing(true);
      setError(null);

      await adminApi.unfreezeWallet(selectedWallet.id);

      // Reload wallets
      const walletsData = await adminApi.getUserWallets(userId!);
      setWallets(walletsData);

      setShowUnfreezeModal(false);
      setSelectedWallet(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to unfreeze wallet');
    } finally {
      setIsProcessing(false);
    }
  };

  const confirmCloseWallet = async () => {
    if (!selectedWallet || !actionReason.trim()) {
      setError('Please provide a reason for closing this wallet');
      return;
    }

    if (actionReason.length < 10) {
      setError('Reason must be at least 10 characters');
      return;
    }

    // Validate balance is zero
    if (selectedWallet.balance > 0) {
      setError(`Cannot close wallet with non-zero balance. Current balance: ${selectedWallet.currency} ${(selectedWallet.balance / 100).toFixed(2)}`);
      return;
    }

    try {
      setIsProcessing(true);
      setError(null);

      await adminApi.closeWallet(selectedWallet.id, { reason: actionReason });

      // Reload wallets
      const walletsData = await adminApi.getUserWallets(userId!);
      setWallets(walletsData);

      setShowCloseModal(false);
      setSelectedWallet(null);
      setActionReason('');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to close wallet');
    } finally {
      setIsProcessing(false);
    }
  };

  const confirmSuspendUser = async () => {
    if (!suspensionReason.trim()) {
      setError('Please provide a reason for suspension');
      return;
    }

    if (suspensionReason.length < 10) {
      setError('Reason must be at least 10 characters');
      return;
    }

    try {
      setIsProcessing(true);
      setError(null);

      await adminApi.suspendUser(userId!, suspensionReason);

      // Reload user data
      await loadUserData();

      setShowSuspendModal(false);
      setSuspensionReason('');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to suspend user');
    } finally {
      setIsProcessing(false);
    }
  };

  const confirmUnsuspendUser = async () => {
    try {
      setIsProcessing(true);
      setError(null);

      await adminApi.unsuspendUser(userId!);

      // Reload user data
      await loadUserData();

      setShowUnsuspendModal(false);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to unsuspend user');
    } finally {
      setIsProcessing(false);
    }
  };

  const tabs: { id: Tab; label: string }[] = [
    { id: 'profile', label: 'Profile' },
    { id: 'kyc', label: 'KYC Details' },
    { id: 'wallets', label: 'Wallets' },
    { id: 'transactions', label: 'Transactions' },
  ];

  if (isLoading) {
    return (
      <AdminLayout title="User Details">
        <div className="space-y-6">
          {/* User Header Skeleton */}
          <Card>
            <div className="flex justify-between items-start">
              <div className="flex-1">
                <div className="flex items-center gap-3 mb-2">
                  <Skeleton className="h-8 w-48" />
                  <Skeleton className="h-6 w-20" />
                </div>
                <Skeleton className="h-4 w-56 mb-1" />
                <Skeleton className="h-4 w-40 mb-1" />
                <Skeleton className="h-3 w-72" />
              </div>
              <div className="flex flex-col gap-2">
                <Skeleton className="h-9 w-32" />
                <Skeleton className="h-9 w-32" />
              </div>
            </div>
          </Card>

          {/* Tab Skeleton */}
          <Skeleton className="h-12 w-full" />

          {/* Content Skeleton */}
          <Card>
            <Skeleton className="h-6 w-40 mb-4" />
            <div className="grid grid-cols-2 gap-4">
              <Skeleton className="h-16" />
              <Skeleton className="h-16" />
              <Skeleton className="h-16" />
              <Skeleton className="h-16" />
            </div>
          </Card>
        </div>
      </AdminLayout>
    );
  }

  if (error && !user) {
    return (
      <AdminLayout title="User Details">
        <Card className="max-w-md mx-auto text-center py-12">
          <div className="w-16 h-16 mx-auto mb-4 rounded-full bg-[var(--color-error-50)] flex items-center justify-center">
            <svg className="w-8 h-8 text-[var(--color-error-600)]" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
            </svg>
          </div>
          <CardTitle className="mb-2">Error Loading User</CardTitle>
          <p className="text-[var(--text-muted)] mb-6">{error || 'User not found'}</p>
          <Button onClick={() => navigate('/')}>
            Back to Dashboard
          </Button>
        </Card>
      </AdminLayout>
    );
  }

  if (!user) return null;

  const breadcrumbs = [
    { label: 'Dashboard', href: '/' },
    { label: 'Users', href: '/users' },
    { label: user?.full_name || 'User Details' },
  ];

  return (
    <AdminLayout title="User Details" breadcrumbs={breadcrumbs}>
      <div className="space-y-6">
        {/* Error Alert */}
        {error && (
          <Alert variant="error" onDismiss={() => setError(null)}>
            {error}
          </Alert>
        )}

        {/* User Header Card */}
        <Card>
          <div className="flex justify-between items-start">
            <div className="flex-1">
              <div className="flex items-center gap-3 mb-2">
                <h2 className="text-2xl font-bold text-[var(--text-primary)]">{user.full_name}</h2>
                <Badge variant={getStatusVariant(user.status)}>
                  {user.status}
                </Badge>
                {user.kyc && (
                  <Badge variant={getKYCStatusVariant(user.kyc.status)}>
                    KYC: {user.kyc.status}
                  </Badge>
                )}
              </div>
              <div className="space-y-1 text-[var(--text-secondary)]">
                <p className="text-sm">{user.email}</p>
                <p className="text-sm">{user.phone}</p>
                <p className="text-xs text-[var(--text-muted)] font-mono">User ID: {user.id}</p>
                <p className="text-xs text-[var(--text-muted)]">
                  Registered: {new Date(user.created_at).toLocaleDateString()}
                </p>
              </div>
            </div>
            <div className="flex flex-col gap-2">
              {user.status === 'suspended' ? (
                <Button
                  onClick={() => setShowUnsuspendModal(true)}
                  size="sm"
                  className="bg-[var(--color-success-600)] hover:bg-[var(--color-success-700)]"
                >
                  Unsuspend User
                </Button>
              ) : user.status !== 'closed' && (
                <Button
                  onClick={() => setShowSuspendModal(true)}
                  size="sm"
                  className="bg-[var(--color-error-600)] hover:bg-[var(--color-error-700)]"
                >
                  Suspend User
                </Button>
              )}
              <Button variant="secondary" size="sm" disabled>
                Reset Password
              </Button>
            </div>
          </div>
        </Card>

        {/* Navigation Tabs */}
        <div className="border-b border-[var(--border-subtle)]">
          <nav className="flex gap-8">
            {tabs.map(tab => (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id)}
                className={cn(
                  'py-4 px-1 border-b-2 font-medium text-sm transition-colors',
                  activeTab === tab.id
                    ? 'border-[var(--interactive-primary)] text-[var(--interactive-primary)]'
                    : 'border-transparent text-[var(--text-muted)] hover:text-[var(--text-primary)] hover:border-[var(--border-default)]'
                )}
              >
                {tab.label}
              </button>
            ))}
          </nav>
        </div>

        {/* Tab Content */}
        {/* Profile Tab */}
        {activeTab === 'profile' && (
          <Card>
            <CardTitle className="mb-4">Profile Information</CardTitle>
            <div className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-[var(--text-muted)] mb-1">Full Name</label>
                  <p className="text-[var(--text-primary)]">{user.full_name}</p>
                </div>
                <div>
                  <label className="block text-sm font-medium text-[var(--text-muted)] mb-1">Email</label>
                  <p className="text-[var(--text-primary)]">{user.email}</p>
                </div>
                <div>
                  <label className="block text-sm font-medium text-[var(--text-muted)] mb-1">Phone</label>
                  <p className="text-[var(--text-primary)]">{user.phone}</p>
                </div>
                <div>
                  <label className="block text-sm font-medium text-[var(--text-muted)] mb-1">Status</label>
                  <Badge variant={getStatusVariant(user.status)}>
                    {user.status}
                  </Badge>
                </div>
                <div>
                  <label className="block text-sm font-medium text-[var(--text-muted)] mb-1">User ID</label>
                  <p className="text-[var(--text-primary)] font-mono text-xs">{user.id}</p>
                </div>
                <div>
                  <label className="block text-sm font-medium text-[var(--text-muted)] mb-1">Registered</label>
                  <p className="text-[var(--text-primary)]">{new Date(user.created_at).toLocaleString()}</p>
                </div>
                <div>
                  <label className="block text-sm font-medium text-[var(--text-muted)] mb-1">Last Updated</label>
                  <p className="text-[var(--text-primary)]">{new Date(user.updated_at).toLocaleString()}</p>
                </div>
              </div>

              {/* Suspension Info */}
              {user.status === 'suspended' && user.suspended_at && (
                <div className="mt-6 pt-6 border-t border-[var(--border-subtle)]">
                  <h4 className="text-md font-semibold text-[var(--color-error-700)] mb-3">Suspension Information</h4>
                  <div className="bg-[var(--color-error-50)] rounded-lg p-4 space-y-2">
                    <div>
                      <label className="block text-sm font-medium text-[var(--color-error-800)] mb-1">Suspended At</label>
                      <p className="text-[var(--color-error-900)]">{new Date(user.suspended_at).toLocaleString()}</p>
                    </div>
                    {user.suspension_reason && (
                      <div>
                        <label className="block text-sm font-medium text-[var(--color-error-800)] mb-1">Reason</label>
                        <p className="text-[var(--color-error-900)]">{user.suspension_reason}</p>
                      </div>
                    )}
                    {user.suspended_by && (
                      <div>
                        <label className="block text-sm font-medium text-[var(--color-error-800)] mb-1">Suspended By</label>
                        <p className="text-[var(--color-error-900)] font-mono text-xs">{user.suspended_by}</p>
                      </div>
                    )}
                  </div>
                </div>
              )}
            </div>
          </Card>
        )}

        {/* KYC Tab */}
        {activeTab === 'kyc' && (
          <Card>
            <CardTitle className="mb-4">KYC Information</CardTitle>
            {user.kyc ? (
              <div className="space-y-4">
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="block text-sm font-medium text-[var(--text-muted)] mb-1">Status</label>
                    <Badge variant={getKYCStatusVariant(user.kyc.status)}>
                      {user.kyc.status}
                    </Badge>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-[var(--text-muted)] mb-1">PAN</label>
                    <p className="text-[var(--text-primary)] font-mono">{user.kyc.pan || 'Not provided'}</p>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-[var(--text-muted)] mb-1">Aadhaar</label>
                    <p className="text-[var(--text-primary)] font-mono">{user.kyc.aadhaar || 'Not provided'}</p>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-[var(--text-muted)] mb-1">Date of Birth</label>
                    <p className="text-[var(--text-primary)]">
                      {user.kyc.date_of_birth ? new Date(user.kyc.date_of_birth).toLocaleDateString() : 'Not provided'}
                    </p>
                  </div>
                  <div className="col-span-2">
                    <label className="block text-sm font-medium text-[var(--text-muted)] mb-1">Address</label>
                    {user.kyc.address ? (
                      <p className="text-[var(--text-primary)]">
                        {user.kyc.address.street}, {user.kyc.address.city}, {user.kyc.address.state} - {user.kyc.address.pin}, {user.kyc.address.country}
                      </p>
                    ) : (
                      <p className="text-[var(--text-muted)]">Not provided</p>
                    )}
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-[var(--text-muted)] mb-1">Submitted</label>
                    <p className="text-[var(--text-primary)]">{new Date(user.kyc.created_at).toLocaleString()}</p>
                  </div>
                  {user.kyc.verified_at && (
                    <div>
                      <label className="block text-sm font-medium text-[var(--text-muted)] mb-1">Verified At</label>
                      <p className="text-[var(--text-primary)]">{new Date(user.kyc.verified_at).toLocaleString()}</p>
                    </div>
                  )}
                  {user.kyc.rejected_at && (
                    <div className="col-span-2">
                      <label className="block text-sm font-medium text-[var(--text-muted)] mb-1">Rejected At</label>
                      <p className="text-[var(--text-primary)] mb-2">{new Date(user.kyc.rejected_at).toLocaleString()}</p>
                      <label className="block text-sm font-medium text-[var(--text-muted)] mb-1">Rejection Reason</label>
                      <p className="text-[var(--color-error-600)]">{user.kyc.rejection_reason}</p>
                    </div>
                  )}
                </div>

                {user.kyc.status === 'pending' && (
                  <div className="mt-6 pt-6 border-t border-[var(--border-subtle)]">
                    <Button onClick={() => navigate('/kyc')}>
                      Review KYC
                    </Button>
                  </div>
                )}
              </div>
            ) : (
              <div className="text-center py-12">
                <div className="w-16 h-16 mx-auto mb-4 rounded-full bg-[var(--surface-secondary)] flex items-center justify-center">
                  <svg className="w-8 h-8 text-[var(--text-muted)]" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                  </svg>
                </div>
                <p className="text-[var(--text-muted)]">No KYC information submitted yet</p>
              </div>
            )}
          </Card>
        )}

        {/* Wallets Tab */}
        {activeTab === 'wallets' && (
          <div>
            {wallets.length > 0 ? (
              <div className="space-y-4">
                {wallets.map((wallet) => (
                  <Card key={wallet.id}>
                    <div className="flex justify-between items-start">
                      <div className="flex-1">
                        <h4 className="text-lg font-semibold text-[var(--text-primary)] mb-2">
                          {wallet.type.toUpperCase()} Wallet
                        </h4>
                        <div className="space-y-1 text-sm">
                          <p className="text-[var(--text-secondary)]">
                            Balance: <span className="font-semibold text-[var(--text-primary)]">
                              {wallet.currency} {(wallet.balance / 100).toFixed(2)}
                            </span>
                          </p>
                          <p className="text-[var(--text-secondary)]">
                            Available: <span className="font-semibold text-[var(--text-primary)]">
                              {wallet.currency} {(wallet.available_balance / 100).toFixed(2)}
                            </span>
                          </p>
                          <p className="text-xs text-[var(--text-muted)] font-mono">Wallet ID: {wallet.id}</p>
                        </div>
                      </div>
                      <div className="flex flex-col gap-2 items-end">
                        <Badge variant={getWalletStatusVariant(wallet.status)}>
                          {wallet.status}
                        </Badge>
                        {wallet.status === 'active' && (
                          <>
                            <Button
                              onClick={(e) => {
                                e.stopPropagation();
                                handleFreezeWallet(wallet);
                              }}
                              variant="secondary"
                              size="sm"
                            >
                              Freeze
                            </Button>
                            <Button
                              onClick={(e) => {
                                e.stopPropagation();
                                handleCloseWallet(wallet);
                              }}
                              disabled={wallet.balance > 0}
                              variant="secondary"
                              size="sm"
                              className="text-[var(--color-error-600)] hover:bg-[var(--color-error-50)]"
                              title={wallet.balance > 0 ? 'Cannot close wallet with non-zero balance' : 'Close wallet permanently'}
                            >
                              Close
                            </Button>
                          </>
                        )}
                        {wallet.status === 'frozen' && (
                          <Button
                            onClick={(e) => {
                              e.stopPropagation();
                              handleUnfreezeWallet(wallet);
                            }}
                            variant="secondary"
                            size="sm"
                          >
                            Unfreeze
                          </Button>
                        )}
                      </div>
                    </div>
                  </Card>
                ))}
              </div>
            ) : (
              <Card className="text-center py-12">
                <div className="w-16 h-16 mx-auto mb-4 rounded-full bg-[var(--surface-secondary)] flex items-center justify-center">
                  <svg className="w-8 h-8 text-[var(--text-muted)]" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 10h18M7 15h1m4 0h1m-7 4h12a3 3 0 003-3V8a3 3 0 00-3-3H6a3 3 0 00-3 3v8a3 3 0 003 3z" />
                  </svg>
                </div>
                <p className="text-[var(--text-muted)]">No wallets found</p>
                <p className="text-sm text-[var(--text-muted)] mt-2">Wallets are created automatically upon KYC verification</p>
              </Card>
            )}
          </div>
        )}

        {/* Transactions Tab */}
        {activeTab === 'transactions' && (
          <div>
            {isLoadingTransactions ? (
              <div className="space-y-4">
                {[1, 2, 3].map(i => (
                  <Card key={i}>
                    <div className="flex justify-between items-start">
                      <div className="flex-1">
                        <div className="flex items-center gap-3 mb-2">
                          <Skeleton className="h-6 w-24" />
                          <Skeleton className="h-6 w-20" />
                        </div>
                        <Skeleton className="h-8 w-32 mb-2" />
                        <Skeleton className="h-4 w-48 mb-1" />
                        <Skeleton className="h-3 w-36" />
                      </div>
                    </div>
                  </Card>
                ))}
              </div>
            ) : transactions.length > 0 ? (
              <div className="space-y-4">
                <p className="text-sm text-[var(--text-muted)]">
                  {transactions.length} transaction{transactions.length !== 1 ? 's' : ''}
                </p>
                {transactions.map((tx) => (
                  <Card
                    key={tx.id}
                    className="cursor-pointer hover:shadow-md transition-shadow"
                    onClick={() => setSelectedTransactionId(tx.id)}
                  >
                    <div className="flex justify-between items-start">
                      <div className="flex-1">
                        <div className="flex items-center gap-3 mb-2">
                          <Badge variant={getTransactionTypeVariant(tx.type)}>
                            {tx.type}
                          </Badge>
                          <Badge variant={getTransactionStatusVariant(tx.status)}>
                            {tx.status}
                          </Badge>
                        </div>
                        <div className="space-y-1 text-sm">
                          <p className="text-xl font-bold text-[var(--text-primary)]">
                            {tx.currency} {(tx.amount / 100).toFixed(2)}
                          </p>
                          <p className="text-[var(--text-secondary)]">{tx.description}</p>
                          <p className="text-xs text-[var(--text-muted)]">
                            {new Date(tx.created_at).toLocaleString()}
                          </p>
                          <p className="text-xs text-[var(--text-muted)] font-mono">TX ID: {tx.id}</p>
                        </div>
                      </div>
                      <Button variant="secondary" size="sm">View</Button>
                    </div>
                  </Card>
                ))}
              </div>
            ) : (
              <Card className="text-center py-12">
                <div className="w-16 h-16 mx-auto mb-4 rounded-full bg-[var(--surface-secondary)] flex items-center justify-center">
                  <svg className="w-8 h-8 text-[var(--text-muted)]" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
                  </svg>
                </div>
                <p className="text-[var(--text-muted)]">No transactions found</p>
              </Card>
            )}
          </div>
        )}
      </div>

      {/* Freeze Wallet Modal */}
      {showFreezeModal && selectedWallet && (
        <div
          className="fixed inset-0 bg-[var(--surface-overlay)] flex items-center justify-center z-50 p-4"
          role="dialog"
          aria-modal="true"
          aria-labelledby="freeze-modal-title"
        >
          <Card className="max-w-md w-full">
            <CardTitle id="freeze-modal-title" className="mb-4">Freeze Wallet</CardTitle>
            <p className="text-sm text-[var(--text-secondary)] mb-4">
              You are about to freeze {selectedWallet.type.toUpperCase()} wallet ({selectedWallet.currency}).
              This will prevent all transactions on this wallet.
            </p>
            <FormField
              label="Reason for freezing (minimum 10 characters)"
              htmlFor="freeze-reason"
            >
              <textarea
                id="freeze-reason"
                value={actionReason}
                onChange={(e) => setActionReason(e.target.value)}
                className={cn(
                  'w-full px-4 py-3 rounded-lg border transition-colors resize-none',
                  'bg-[var(--surface-input)] border-[var(--border-default)]',
                  'text-[var(--text-primary)] placeholder:text-[var(--text-muted)]',
                  'focus:outline-none focus:ring-2 focus:ring-[var(--interactive-primary)] focus:border-transparent'
                )}
                rows={3}
                placeholder="Enter reason for freezing this wallet..."
              />
            </FormField>
            <div className="flex gap-3 mt-4">
              <Button
                variant="secondary"
                onClick={() => {
                  setShowFreezeModal(false);
                  setSelectedWallet(null);
                  setActionReason('');
                }}
                disabled={isProcessing}
                className="flex-1"
              >
                Cancel
              </Button>
              <Button
                onClick={confirmFreezeWallet}
                disabled={isProcessing || !actionReason.trim() || actionReason.length < 10}
                loading={isProcessing}
                className="flex-1 bg-[var(--color-warning-600)] hover:bg-[var(--color-warning-700)]"
              >
                Freeze Wallet
              </Button>
            </div>
          </Card>
        </div>
      )}

      {/* Unfreeze Wallet Modal */}
      {showUnfreezeModal && selectedWallet && (
        <div
          className="fixed inset-0 bg-[var(--surface-overlay)] flex items-center justify-center z-50 p-4"
          role="dialog"
          aria-modal="true"
          aria-labelledby="unfreeze-modal-title"
        >
          <Card className="max-w-md w-full">
            <CardTitle id="unfreeze-modal-title" className="mb-4">Unfreeze Wallet</CardTitle>
            <p className="text-sm text-[var(--text-secondary)] mb-6">
              You are about to unfreeze {selectedWallet.type.toUpperCase()} wallet ({selectedWallet.currency}).
              This will restore normal transaction capabilities.
            </p>
            <div className="flex gap-3">
              <Button
                variant="secondary"
                onClick={() => {
                  setShowUnfreezeModal(false);
                  setSelectedWallet(null);
                }}
                disabled={isProcessing}
                className="flex-1"
              >
                Cancel
              </Button>
              <Button
                onClick={confirmUnfreezeWallet}
                disabled={isProcessing}
                loading={isProcessing}
                className="flex-1 bg-[var(--color-success-600)] hover:bg-[var(--color-success-700)]"
              >
                Unfreeze Wallet
              </Button>
            </div>
          </Card>
        </div>
      )}

      {/* Close Wallet Modal */}
      {showCloseModal && selectedWallet && (
        <div
          className="fixed inset-0 bg-[var(--surface-overlay)] flex items-center justify-center z-50 p-4"
          role="dialog"
          aria-modal="true"
          aria-labelledby="close-modal-title"
        >
          <Card className="max-w-md w-full">
            <CardTitle id="close-modal-title" className="mb-4">Close Wallet</CardTitle>
            <Alert variant="error" className="mb-4">
              <strong>Warning: This action is permanent!</strong>
              <p className="text-sm mt-1">
                Closing a wallet is irreversible. Ensure the balance is zero before proceeding.
              </p>
            </Alert>
            <p className="text-sm text-[var(--text-secondary)] mb-2">
              Wallet: {selectedWallet.type.toUpperCase()} ({selectedWallet.currency})
            </p>
            <p className="text-sm text-[var(--text-secondary)] mb-4">
              Balance: {selectedWallet.currency} {(selectedWallet.balance / 100).toFixed(2)}
            </p>
            <FormField
              label="Reason for closing (minimum 10 characters)"
              htmlFor="close-reason"
            >
              <textarea
                id="close-reason"
                value={actionReason}
                onChange={(e) => setActionReason(e.target.value)}
                className={cn(
                  'w-full px-4 py-3 rounded-lg border transition-colors resize-none',
                  'bg-[var(--surface-input)] border-[var(--border-default)]',
                  'text-[var(--text-primary)] placeholder:text-[var(--text-muted)]',
                  'focus:outline-none focus:ring-2 focus:ring-[var(--interactive-primary)] focus:border-transparent'
                )}
                rows={3}
                placeholder="Enter reason for closing this wallet..."
              />
            </FormField>
            <div className="flex gap-3 mt-4">
              <Button
                variant="secondary"
                onClick={() => {
                  setShowCloseModal(false);
                  setSelectedWallet(null);
                  setActionReason('');
                }}
                disabled={isProcessing}
                className="flex-1"
              >
                Cancel
              </Button>
              <Button
                onClick={confirmCloseWallet}
                disabled={isProcessing || !actionReason.trim() || actionReason.length < 10}
                loading={isProcessing}
                className="flex-1 bg-[var(--color-error-600)] hover:bg-[var(--color-error-700)]"
              >
                Close Wallet
              </Button>
            </div>
          </Card>
        </div>
      )}

      {/* Suspend User Modal */}
      {showSuspendModal && (
        <div
          className="fixed inset-0 bg-[var(--surface-overlay)] flex items-center justify-center z-50 p-4"
          role="dialog"
          aria-modal="true"
          aria-labelledby="suspend-modal-title"
        >
          <Card className="max-w-md w-full">
            <CardTitle id="suspend-modal-title" className="mb-4">Suspend User</CardTitle>
            <Alert variant="warning" className="mb-4">
              Suspending this user will prevent them from accessing their account and performing any transactions.
            </Alert>
            <p className="text-sm text-[var(--text-secondary)] mb-4">
              User: {user?.full_name} ({user?.email})
            </p>
            <FormField
              label="Reason for suspension (minimum 10 characters)"
              htmlFor="suspension-reason"
              required
            >
              <textarea
                id="suspension-reason"
                value={suspensionReason}
                onChange={(e) => setSuspensionReason(e.target.value)}
                className={cn(
                  'w-full px-4 py-3 rounded-lg border transition-colors resize-none',
                  'bg-[var(--surface-input)] border-[var(--border-default)]',
                  'text-[var(--text-primary)] placeholder:text-[var(--text-muted)]',
                  'focus:outline-none focus:ring-2 focus:ring-[var(--interactive-primary)] focus:border-transparent'
                )}
                rows={3}
                placeholder="Enter detailed reason for suspending this user..."
              />
            </FormField>
            <div className="flex gap-3 mt-4">
              <Button
                variant="secondary"
                onClick={() => {
                  setShowSuspendModal(false);
                  setSuspensionReason('');
                }}
                disabled={isProcessing}
                className="flex-1"
              >
                Cancel
              </Button>
              <Button
                onClick={confirmSuspendUser}
                disabled={isProcessing || !suspensionReason.trim() || suspensionReason.length < 10}
                loading={isProcessing}
                className="flex-1 bg-[var(--color-error-600)] hover:bg-[var(--color-error-700)]"
              >
                Suspend User
              </Button>
            </div>
          </Card>
        </div>
      )}

      {/* Unsuspend User Modal */}
      {showUnsuspendModal && (
        <div
          className="fixed inset-0 bg-[var(--surface-overlay)] flex items-center justify-center z-50 p-4"
          role="dialog"
          aria-modal="true"
          aria-labelledby="unsuspend-modal-title"
        >
          <Card className="max-w-md w-full">
            <CardTitle id="unsuspend-modal-title" className="mb-4">Unsuspend User</CardTitle>
            <p className="text-sm text-[var(--text-secondary)] mb-4">
              You are about to unsuspend <span className="font-semibold">{user?.full_name}</span>.
              This will restore full account access and allow them to perform transactions again.
            </p>
            {user?.suspended_at && (
              <div className="bg-[var(--surface-secondary)] rounded-lg p-3 mb-4">
                <p className="text-xs text-[var(--text-secondary)]">Originally suspended:</p>
                <p className="text-sm text-[var(--text-primary)] font-medium">{new Date(user.suspended_at).toLocaleString()}</p>
                {user.suspension_reason && (
                  <>
                    <p className="text-xs text-[var(--text-secondary)] mt-2">Reason:</p>
                    <p className="text-sm text-[var(--text-primary)]">{user.suspension_reason}</p>
                  </>
                )}
              </div>
            )}
            <div className="flex gap-3">
              <Button
                variant="secondary"
                onClick={() => setShowUnsuspendModal(false)}
                disabled={isProcessing}
                className="flex-1"
              >
                Cancel
              </Button>
              <Button
                onClick={confirmUnsuspendUser}
                disabled={isProcessing}
                loading={isProcessing}
                className="flex-1 bg-[var(--color-success-600)] hover:bg-[var(--color-success-700)]"
              >
                Unsuspend User
              </Button>
            </div>
          </Card>
        </div>
      )}

      {/* Transaction Detail Modal */}
      {selectedTransactionId && (
        <TransactionDetailModal
          transactionId={selectedTransactionId}
          onClose={() => setSelectedTransactionId(null)}
        />
      )}
    </AdminLayout>
  );
}
