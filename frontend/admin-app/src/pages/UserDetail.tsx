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
  const [wallets] = useState<Wallet[]>([]);
  const [transactions] = useState<Transaction[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

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

      // Load user details
      const userData = await adminApi.getUserDetails(userId);
      setUser(userData);

      // TODO: Load wallets when endpoint is ready
      // const walletsData = await adminApi.getUserWallets(userId);
      // setWallets(walletsData);

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
                            'bg-gray-100 text-gray-800'
                          }`}>
                            {wallet.status}
                          </span>
                          {wallet.status === 'active' && (
                            <button className="btn-secondary text-sm">Freeze</button>
                          )}
                          {wallet.status === 'frozen' && (
                            <button className="btn-secondary text-sm">Unfreeze</button>
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
    </div>
  );
}
