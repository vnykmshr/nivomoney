/**
 * User Detail Page
 * Comprehensive view of user profile, KYC, wallets, and transactions
 */

import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { adminApi } from '../lib/adminApi';
import type { User, Wallet, Transaction } from '@nivo/shared';

export function UserDetail() {
  const { userId } = useParams<{ userId: string }>();
  const navigate = useNavigate();
  const [activeTab, setActiveTab] = useState<'profile' | 'kyc' | 'wallets' | 'transactions'>('profile');
  const [user, setUser] = useState<User | null>(null);
  const [wallets, setWallets] = useState<Wallet[]>([]);
  const [transactions] = useState<Transaction[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Wallet action modals
  const [showFreezeModal, setShowFreezeModal] = useState(false);
  const [showUnfreezeModal, setShowUnfreezeModal] = useState(false);
  const [showCloseModal, setShowCloseModal] = useState(false);
  const [selectedWallet, setSelectedWallet] = useState<Wallet | null>(null);
  const [actionReason, setActionReason] = useState('');
  const [isProcessing, setIsProcessing] = useState(false);

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

      // TODO: Load transactions when endpoint is ready
      // const txData = await adminApi.getUserTransactions(userId);
      // setTransactions(txData);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load user data');
    } finally {
      setIsLoading(false);
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'active': return 'bg-green-100 text-green-800';
      case 'pending': return 'bg-yellow-100 text-yellow-800';
      case 'suspended': return 'bg-red-100 text-red-800';
      default: return 'bg-gray-100 text-gray-800';
    }
  };

  const getKYCStatusColor = (status?: string) => {
    switch (status) {
      case 'verified': return 'bg-green-100 text-green-800';
      case 'pending': return 'bg-yellow-100 text-yellow-800';
      case 'rejected': return 'bg-red-100 text-red-800';
      default: return 'bg-gray-100 text-gray-800';
    }
  };

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

  if (isLoading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-gray-500">Loading user details...</div>
      </div>
    );
  }

  if (error || !user) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="max-w-md w-full">
          <div className="card">
            <div className="text-center">
              <div className="text-red-600 text-5xl mb-4">⚠️</div>
              <h2 className="text-xl font-semibold text-gray-900 mb-2">Error Loading User</h2>
              <p className="text-gray-600 mb-6">{error || 'User not found'}</p>
              <button onClick={() => navigate('/')} className="btn-primary">
                Back to Dashboard
              </button>
            </div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <nav className="bg-white shadow-sm">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between h-16 items-center">
            <div className="flex items-center space-x-4">
              <button
                onClick={() => navigate('/')}
                className="text-gray-600 hover:text-gray-900"
              >
                ← Back
              </button>
              <h1 className="text-xl font-bold text-primary-600">User Details</h1>
            </div>
          </div>
        </div>
      </nav>

      {/* Main Content */}
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* User Header Card */}
        <div className="card mb-6">
          <div className="flex justify-between items-start">
            <div className="flex-1">
              <div className="flex items-center space-x-3 mb-2">
                <h2 className="text-2xl font-bold text-gray-900">{user.full_name}</h2>
                <span className={`px-3 py-1 rounded-full text-sm ${getStatusColor(user.status)}`}>
                  {user.status}
                </span>
                {user.kyc && (
                  <span className={`px-3 py-1 rounded-full text-sm ${getKYCStatusColor(user.kyc.status)}`}>
                    KYC: {user.kyc.status}
                  </span>
                )}
              </div>
              <div className="space-y-1 text-gray-600">
                <p className="text-sm">📧 {user.email}</p>
                <p className="text-sm">📱 {user.phone}</p>
                <p className="text-xs text-gray-500">User ID: {user.id}</p>
                <p className="text-xs text-gray-500">
                  Registered: {new Date(user.created_at).toLocaleDateString()}
                </p>
              </div>
            </div>
            <div className="flex flex-col gap-2">
              <button className="btn-secondary text-sm">
                Suspend User
              </button>
              <button className="btn-secondary text-sm">
                Reset Password
              </button>
            </div>
          </div>
        </div>

        {/* Navigation Tabs */}
        <div className="mb-6 border-b border-gray-200">
          <nav className="flex space-x-8">
            <button
              onClick={() => setActiveTab('profile')}
              className={`py-4 px-1 border-b-2 font-medium text-sm ${
                activeTab === 'profile'
                  ? 'border-primary-500 text-primary-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              }`}
            >
              Profile
            </button>
            <button
              onClick={() => setActiveTab('kyc')}
              className={`py-4 px-1 border-b-2 font-medium text-sm ${
                activeTab === 'kyc'
                  ? 'border-primary-500 text-primary-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              }`}
            >
              KYC Details
            </button>
            <button
              onClick={() => setActiveTab('wallets')}
              className={`py-4 px-1 border-b-2 font-medium text-sm ${
                activeTab === 'wallets'
                  ? 'border-primary-500 text-primary-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              }`}
            >
              Wallets
            </button>
            <button
              onClick={() => setActiveTab('transactions')}
              className={`py-4 px-1 border-b-2 font-medium text-sm ${
                activeTab === 'transactions'
                  ? 'border-primary-500 text-primary-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              }`}
            >
              Transactions
            </button>
          </nav>
        </div>

        {/* Tab Content */}
        <div>
          {/* Profile Tab */}
          {activeTab === 'profile' && (
            <div className="card">
              <h3 className="text-lg font-semibold text-gray-900 mb-4">Profile Information</h3>
              <div className="space-y-4">
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Full Name</label>
                    <p className="text-gray-900">{user.full_name}</p>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Email</label>
                    <p className="text-gray-900">{user.email}</p>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Phone</label>
                    <p className="text-gray-900">{user.phone}</p>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Status</label>
                    <span className={`inline-block px-3 py-1 rounded-full text-sm ${getStatusColor(user.status)}`}>
                      {user.status}
                    </span>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">User ID</label>
                    <p className="text-gray-900 font-mono text-xs">{user.id}</p>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Registered</label>
                    <p className="text-gray-900">{new Date(user.created_at).toLocaleString()}</p>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Last Updated</label>
                    <p className="text-gray-900">{new Date(user.updated_at).toLocaleString()}</p>
                  </div>
                </div>
              </div>
            </div>
          )}

          {/* KYC Tab */}
          {activeTab === 'kyc' && (
            <div className="card">
              <h3 className="text-lg font-semibold text-gray-900 mb-4">KYC Information</h3>
              {user.kyc ? (
                <div className="space-y-4">
                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">Status</label>
                      <span className={`inline-block px-3 py-1 rounded-full text-sm ${getKYCStatusColor(user.kyc.status)}`}>
                        {user.kyc.status}
                      </span>
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">PAN</label>
                      <p className="text-gray-900 font-mono">{user.kyc.pan || 'Not provided'}</p>
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">Aadhaar</label>
                      <p className="text-gray-900 font-mono">{user.kyc.aadhaar || 'Not provided'}</p>
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">Date of Birth</label>
                      <p className="text-gray-900">
                        {user.kyc.date_of_birth ? new Date(user.kyc.date_of_birth).toLocaleDateString() : 'Not provided'}
                      </p>
                    </div>
                    <div className="col-span-2">
                      <label className="block text-sm font-medium text-gray-700 mb-1">Address</label>
                      {user.kyc.address ? (
                        <p className="text-gray-900">
                          {user.kyc.address.street}, {user.kyc.address.city}, {user.kyc.address.state} - {user.kyc.address.pin}, {user.kyc.address.country}
                        </p>
                      ) : (
                        <p className="text-gray-500">Not provided</p>
                      )}
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">Submitted</label>
                      <p className="text-gray-900">{new Date(user.kyc.created_at).toLocaleString()}</p>
                    </div>
                    {user.kyc.verified_at && (
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">Verified At</label>
                        <p className="text-gray-900">{new Date(user.kyc.verified_at).toLocaleString()}</p>
                      </div>
                    )}
                    {user.kyc.rejected_at && (
                      <div className="col-span-2">
                        <label className="block text-sm font-medium text-gray-700 mb-1">Rejected At</label>
                        <p className="text-gray-900 mb-2">{new Date(user.kyc.rejected_at).toLocaleString()}</p>
                        <label className="block text-sm font-medium text-gray-700 mb-1">Rejection Reason</label>
                        <p className="text-red-600">{user.kyc.rejection_reason}</p>
                      </div>
                    )}
                  </div>

                  {user.kyc.status === 'pending' && (
                    <div className="mt-6 pt-6 border-t border-gray-200">
                      <div className="flex gap-3">
                        <button
                          onClick={() => navigate('/kyc')}
                          className="btn-primary"
                        >
                          Review KYC →
                        </button>
                      </div>
                    </div>
                  )}
                </div>
              ) : (
                <div className="text-center py-12 text-gray-500">
                  <div className="text-5xl mb-4">📋</div>
                  <p>No KYC information submitted yet</p>
                </div>
              )}
            </div>
          )}

          {/* Wallets Tab */}
          {activeTab === 'wallets' && (
            <div>
              {wallets.length > 0 ? (
                <div className="space-y-4">
                  {wallets.map((wallet) => (
                    <div key={wallet.id} className="card">
                      <div className="flex justify-between items-start">
                        <div className="flex-1">
                          <h4 className="text-lg font-semibold text-gray-900 mb-2">
                            {wallet.type.toUpperCase()} Wallet
                          </h4>
                          <div className="space-y-1 text-sm">
                            <p className="text-gray-600">
                              Balance: <span className="font-semibold text-gray-900">
                                {wallet.currency} {(wallet.balance / 100).toFixed(2)}
                              </span>
                            </p>
                            <p className="text-gray-600">
                              Available: <span className="font-semibold text-gray-900">
                                {wallet.currency} {(wallet.available_balance / 100).toFixed(2)}
                              </span>
                            </p>
                            <p className="text-xs text-gray-500">Wallet ID: {wallet.id}</p>
                          </div>
                        </div>
                        <div className="flex flex-col gap-2">
                          <span className={`px-3 py-1 rounded-full text-sm ${
                            wallet.status === 'active' ? 'bg-green-100 text-green-800' :
                            wallet.status === 'frozen' ? 'bg-yellow-100 text-yellow-800' :
                            wallet.status === 'closed' ? 'bg-red-100 text-red-800' :
                            'bg-gray-100 text-gray-800'
                          }`}>
                            {wallet.status}
                          </span>
                          {wallet.status === 'active' && (
                            <>
                              <button
                                onClick={(e) => {
                                  e.stopPropagation();
                                  handleFreezeWallet(wallet);
                                }}
                                className="btn-secondary text-sm"
                              >
                                Freeze
                              </button>
                              <button
                                onClick={(e) => {
                                  e.stopPropagation();
                                  handleCloseWallet(wallet);
                                }}
                                className="btn-secondary text-sm text-red-600 hover:bg-red-50"
                              >
                                Close
                              </button>
                            </>
                          )}
                          {wallet.status === 'frozen' && (
                            <button
                              onClick={(e) => {
                                e.stopPropagation();
                                handleUnfreezeWallet(wallet);
                              }}
                              className="btn-secondary text-sm"
                            >
                              Unfreeze
                            </button>
                          )}
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              ) : (
                <div className="card text-center py-12 text-gray-500">
                  <div className="text-5xl mb-4">💰</div>
                  <p>No wallets found</p>
                  <p className="text-sm mt-2">Wallets are created automatically upon KYC verification</p>
                </div>
              )}
            </div>
          )}

          {/* Transactions Tab */}
          {activeTab === 'transactions' && (
            <div>
              {transactions.length > 0 ? (
                <div className="space-y-4">
                  {transactions.map((tx) => (
                    <div key={tx.id} className="card">
                      <div className="flex justify-between items-start">
                        <div className="flex-1">
                          <div className="flex items-center space-x-3 mb-2">
                            <h4 className="text-lg font-semibold text-gray-900">
                              {tx.type.toUpperCase()}
                            </h4>
                            <span className={`px-2 py-1 rounded-full text-xs ${
                              tx.status === 'completed' ? 'bg-green-100 text-green-800' :
                              tx.status === 'pending' ? 'bg-yellow-100 text-yellow-800' :
                              tx.status === 'failed' ? 'bg-red-100 text-red-800' :
                              'bg-gray-100 text-gray-800'
                            }`}>
                              {tx.status}
                            </span>
                          </div>
                          <div className="space-y-1 text-sm">
                            <p className="text-gray-900 font-semibold">
                              {tx.currency} {(tx.amount / 100).toFixed(2)}
                            </p>
                            <p className="text-gray-600">{tx.description}</p>
                            <p className="text-xs text-gray-500">
                              {new Date(tx.created_at).toLocaleString()}
                            </p>
                            <p className="text-xs text-gray-500">TX ID: {tx.id}</p>
                          </div>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              ) : (
                <div className="card text-center py-12 text-gray-500">
                  <div className="text-5xl mb-4">📊</div>
                  <p>No transactions found</p>
                </div>
              )}
            </div>
          )}
        </div>
      </div>

      {/* Freeze Wallet Modal */}
      {showFreezeModal && selectedWallet && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 max-w-md w-full mx-4">
            <h3 className="text-lg font-semibold text-gray-900 mb-4">Freeze Wallet</h3>
            <p className="text-sm text-gray-600 mb-4">
              You are about to freeze {selectedWallet.type.toUpperCase()} wallet ({selectedWallet.currency}).
              This will prevent all transactions on this wallet.
            </p>
            <div className="mb-4">
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Reason for freezing (minimum 10 characters)
              </label>
              <textarea
                value={actionReason}
                onChange={(e) => setActionReason(e.target.value)}
                className="input-field w-full"
                rows={3}
                placeholder="Enter reason for freezing this wallet..."
              />
            </div>
            <div className="flex gap-3">
              <button
                onClick={() => {
                  setShowFreezeModal(false);
                  setSelectedWallet(null);
                  setActionReason('');
                }}
                disabled={isProcessing}
                className="btn-secondary flex-1"
              >
                Cancel
              </button>
              <button
                onClick={confirmFreezeWallet}
                disabled={isProcessing || !actionReason.trim() || actionReason.length < 10}
                className="btn-primary flex-1 bg-yellow-600 hover:bg-yellow-700"
              >
                {isProcessing ? 'Freezing...' : 'Freeze Wallet'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Unfreeze Wallet Modal */}
      {showUnfreezeModal && selectedWallet && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 max-w-md w-full mx-4">
            <h3 className="text-lg font-semibold text-gray-900 mb-4">Unfreeze Wallet</h3>
            <p className="text-sm text-gray-600 mb-6">
              You are about to unfreeze {selectedWallet.type.toUpperCase()} wallet ({selectedWallet.currency}).
              This will restore normal transaction capabilities.
            </p>
            <div className="flex gap-3">
              <button
                onClick={() => {
                  setShowUnfreezeModal(false);
                  setSelectedWallet(null);
                }}
                disabled={isProcessing}
                className="btn-secondary flex-1"
              >
                Cancel
              </button>
              <button
                onClick={confirmUnfreezeWallet}
                disabled={isProcessing}
                className="btn-primary flex-1 bg-green-600 hover:bg-green-700"
              >
                {isProcessing ? 'Unfreezing...' : 'Unfreeze Wallet'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Close Wallet Modal */}
      {showCloseModal && selectedWallet && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 max-w-md w-full mx-4">
            <h3 className="text-lg font-semibold text-gray-900 mb-4">Close Wallet</h3>
            <div className="bg-red-50 border border-red-200 rounded-lg p-4 mb-4">
              <p className="text-sm text-red-800 font-medium">⚠️ Warning: This action is permanent!</p>
              <p className="text-sm text-red-700 mt-1">
                Closing a wallet is irreversible. Ensure the balance is zero before proceeding.
              </p>
            </div>
            <p className="text-sm text-gray-600 mb-2">
              Wallet: {selectedWallet.type.toUpperCase()} ({selectedWallet.currency})
            </p>
            <p className="text-sm text-gray-600 mb-4">
              Balance: {selectedWallet.currency} {(selectedWallet.balance / 100).toFixed(2)}
            </p>
            <div className="mb-4">
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Reason for closing (minimum 10 characters)
              </label>
              <textarea
                value={actionReason}
                onChange={(e) => setActionReason(e.target.value)}
                className="input-field w-full"
                rows={3}
                placeholder="Enter reason for closing this wallet..."
              />
            </div>
            <div className="flex gap-3">
              <button
                onClick={() => {
                  setShowCloseModal(false);
                  setSelectedWallet(null);
                  setActionReason('');
                }}
                disabled={isProcessing}
                className="btn-secondary flex-1"
              >
                Cancel
              </button>
              <button
                onClick={confirmCloseWallet}
                disabled={isProcessing || !actionReason.trim() || actionReason.length < 10}
                className="btn-primary flex-1 bg-red-600 hover:bg-red-700"
              >
                {isProcessing ? 'Closing...' : 'Close Wallet'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
