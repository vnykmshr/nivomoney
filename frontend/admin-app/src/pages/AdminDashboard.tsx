/**
 * Admin Dashboard - Central Hub for Operations
 * Migrated from user-app to admin-app
 *
 * Primary workflow: Notification-driven admin actions
 * - View pending KYC reviews
 * - Manage users
 * - Monitor transactions
 * - View system statistics
 */

import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAdminAuthStore } from '../stores/adminAuthStore';
import { adminApi } from '../lib/adminApi';
import type { User, AdminStats } from '@nivo/shared';

interface NotificationItem {
  id: string;
  type: 'kyc_review' | 'large_transaction' | 'user_report';
  title: string;
  description: string;
  priority: 'high' | 'medium' | 'low';
  link: string;
  createdAt: string;
}

export function AdminDashboard() {
  const navigate = useNavigate();
  const { user, logout } = useAdminAuthStore();
  const [activeTab, setActiveTab] = useState<'notifications' | 'users' | 'transactions' | 'stats'>('notifications');
  const [stats, setStats] = useState<AdminStats | null>(null);
  const [notifications, setNotifications] = useState<NotificationItem[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // User search
  const [searchQuery, setSearchQuery] = useState('');
  const [searchResults, setSearchResults] = useState<User[]>([]);
  const [isSearching, setIsSearching] = useState(false);

  // Transaction search
  const [txSearchQuery, setTxSearchQuery] = useState('');

  useEffect(() => {
    loadDashboardData();
  }, []);

  const loadDashboardData = async () => {
    try {
      setIsLoading(true);

      // Fetch admin stats and pending KYCs in parallel
      const [statsData, pendingKYCs] = await Promise.all([
        adminApi.getAdminStats(),
        adminApi.listPendingKYCs(),
      ]);

      // Convert to notification items
      const kycNotifications: NotificationItem[] = pendingKYCs.map((item) => ({
        id: item.user.id,
        type: 'kyc_review' as const,
        title: `KYC Review: ${item.user.full_name}`,
        description: `Submitted on ${new Date(item.kyc.created_at).toLocaleDateString()}`,
        priority: 'high' as const,
        link: `/kyc`,
        createdAt: item.kyc.created_at,
      }));

      setNotifications(kycNotifications);
      setStats(statsData);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load dashboard data');
    } finally {
      setIsLoading(false);
    }
  };

  const handleSearch = async () => {
    if (!searchQuery.trim()) {
      setSearchResults([]);
      return;
    }

    setIsSearching(true);
    setError(null);

    try {
      const results = await adminApi.searchUsers(searchQuery);
      setSearchResults(results);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Search failed');
      setSearchResults([]);
    } finally {
      setIsSearching(false);
    }
  };

  const handleLogout = async () => {
    await logout();
    navigate('/login');
  };

  const getPriorityColor = (priority: string) => {
    switch (priority) {
      case 'high': return 'bg-red-100 text-red-800';
      case 'medium': return 'bg-yellow-100 text-yellow-800';
      case 'low': return 'bg-blue-100 text-blue-800';
      default: return 'bg-gray-100 text-gray-800';
    }
  };

  const getNotificationIcon = (type: string) => {
    switch (type) {
      case 'kyc_review':
        return '📋';
      case 'large_transaction':
        return '💰';
      case 'user_report':
        return '⚠️';
      default:
        return '📢';
    }
  };

  if (isLoading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-gray-500">Loading admin dashboard...</div>
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
              <h1 className="text-xl font-bold text-primary-600">Nivo Money Admin</h1>
              <span className="px-2 py-1 bg-purple-100 text-purple-800 text-xs rounded-full">
                Admin
              </span>
            </div>
            <div className="flex items-center space-x-4">
              <span className="text-gray-700">{user?.full_name}</span>
              <button onClick={handleLogout} className="btn-secondary">
                Logout
              </button>
            </div>
          </div>
        </div>
      </nav>

      {/* Main Content */}
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Error Alert */}
        {error && (
          <div className="mb-6 p-4 bg-red-100 text-red-800 rounded-lg flex justify-between items-center">
            <span>{error}</span>
            <button onClick={() => setError(null)} className="text-red-600 hover:text-red-800">
              ✕
            </button>
          </div>
        )}

        {/* Stats Cards */}
        <div className="grid grid-cols-1 md:grid-cols-5 gap-4 mb-8">
          <div className="card">
            <div className="text-sm text-gray-600 mb-1">Total Users</div>
            <div className="text-3xl font-bold text-gray-900">{stats?.total_users || '-'}</div>
          </div>
          <div className="card">
            <div className="text-sm text-gray-600 mb-1">Active Users</div>
            <div className="text-3xl font-bold text-green-600">{stats?.active_users || '-'}</div>
          </div>
          <div className="card">
            <div className="text-sm text-gray-600 mb-1">Pending KYC</div>
            <div className="text-3xl font-bold text-orange-600">{stats?.pending_kyc || 0}</div>
          </div>
          <div className="card">
            <div className="text-sm text-gray-600 mb-1">Total Wallets</div>
            <div className="text-3xl font-bold text-blue-600">{stats?.total_wallets || '-'}</div>
          </div>
          <div className="card">
            <div className="text-sm text-gray-600 mb-1">Transactions</div>
            <div className="text-3xl font-bold text-purple-600">{stats?.total_transactions || '-'}</div>
          </div>
        </div>

        {/* Navigation Tabs */}
        <div className="mb-6 border-b border-gray-200">
          <nav className="flex space-x-8">
            <button
              onClick={() => setActiveTab('notifications')}
              className={`py-4 px-1 border-b-2 font-medium text-sm ${
                activeTab === 'notifications'
                  ? 'border-primary-500 text-primary-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              }`}
            >
              Notifications
              {notifications.length > 0 && (
                <span className="ml-2 px-2 py-1 bg-red-100 text-red-800 text-xs rounded-full">
                  {notifications.length}
                </span>
              )}
            </button>
            <button
              onClick={() => setActiveTab('users')}
              className={`py-4 px-1 border-b-2 font-medium text-sm ${
                activeTab === 'users'
                  ? 'border-primary-500 text-primary-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              }`}
            >
              User Management
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
            <button
              onClick={() => setActiveTab('stats')}
              className={`py-4 px-1 border-b-2 font-medium text-sm ${
                activeTab === 'stats'
                  ? 'border-primary-500 text-primary-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              }`}
            >
              Statistics
            </button>
          </nav>
        </div>

        {/* Tab Content */}
        <div>
          {/* Notifications Tab */}
          {activeTab === 'notifications' && (
            <div>
              <h2 className="text-xl font-semibold text-gray-900 mb-4">Pending Actions</h2>

              {notifications.length === 0 ? (
                <div className="card text-center py-12">
                  <div className="text-6xl mb-4">✨</div>
                  <h3 className="text-lg font-medium text-gray-900 mb-2">All Caught Up!</h3>
                  <p className="text-gray-600">No pending notifications at the moment.</p>
                </div>
              ) : (
                <div className="space-y-4">
                  {notifications.map((notification) => (
                    <div
                      key={notification.id}
                      className="card cursor-pointer hover:shadow-md transition-shadow"
                      onClick={() => navigate(notification.link)}
                    >
                      <div className="flex items-start justify-between">
                        <div className="flex items-start space-x-4">
                          <div className="text-3xl">{getNotificationIcon(notification.type)}</div>
                          <div className="flex-1">
                            <h3 className="text-lg font-semibold text-gray-900 mb-1">
                              {notification.title}
                            </h3>
                            <p className="text-sm text-gray-600 mb-2">{notification.description}</p>
                            <div className="flex items-center space-x-3">
                              <span className={`px-2 py-1 text-xs rounded-full ${getPriorityColor(notification.priority)}`}>
                                {notification.priority.toUpperCase()}
                              </span>
                              <span className="text-xs text-gray-500">
                                {new Date(notification.createdAt).toLocaleString()}
                              </span>
                            </div>
                          </div>
                        </div>
                        <button className="btn-primary">
                          Review →
                        </button>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          )}

          {/* Users Tab */}
          {activeTab === 'users' && (
            <div>
              <h2 className="text-xl font-semibold text-gray-900 mb-4">User Management</h2>

              {/* Search Bar */}
              <div className="card mb-6">
                <div className="flex gap-3">
                  <input
                    type="text"
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    onKeyPress={(e) => e.key === 'Enter' && handleSearch()}
                    placeholder="Search by email, phone, or name..."
                    className="input-field flex-1"
                  />
                  <button
                    onClick={handleSearch}
                    disabled={isSearching}
                    className="btn-primary"
                  >
                    {isSearching ? 'Searching...' : 'Search'}
                  </button>
                </div>
              </div>

              {/* Search Results */}
              {searchResults.length > 0 ? (
                <div className="space-y-4">
                  {searchResults.map((user) => (
                    <div
                      key={user.id}
                      className="card cursor-pointer hover:shadow-md transition-shadow"
                      onClick={() => navigate(`/users/${user.id}`)}
                    >
                      <div className="flex justify-between items-start">
                        <div className="flex-1">
                          <h3 className="text-lg font-semibold text-gray-900">{user.full_name}</h3>
                          <p className="text-sm text-gray-600">{user.email}</p>
                          <p className="text-sm text-gray-600">{user.phone}</p>
                          <p className="text-xs text-gray-500 mt-1">
                            User ID: {user.id.slice(0, 8)}...
                          </p>
                        </div>
                        <div className="flex flex-col items-end gap-2">
                          <span className={`px-3 py-1 rounded-full text-sm ${
                            user.status === 'active' ? 'bg-green-100 text-green-800' :
                            user.status === 'pending' ? 'bg-yellow-100 text-yellow-800' :
                            'bg-gray-100 text-gray-800'
                          }`}>
                            {user.status}
                          </span>
                          <button className="btn-primary text-sm">
                            View Details →
                          </button>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              ) : (
                <div className="card text-center py-12 text-gray-500">
                  Search for users by email, phone, or name
                </div>
              )}
            </div>
          )}

          {/* Transactions Tab */}
          {activeTab === 'transactions' && (
            <div>
              <h2 className="text-xl font-semibold text-gray-900 mb-4">Transaction Lookup</h2>

              <div className="card mb-6">
                <div className="flex gap-3">
                  <input
                    type="text"
                    value={txSearchQuery}
                    onChange={(e) => setTxSearchQuery(e.target.value)}
                    placeholder="Search by transaction ID or user email..."
                    className="input-field flex-1"
                  />
                  <button className="btn-primary">
                    Search
                  </button>
                </div>
              </div>

              <div className="card text-center py-12 text-gray-500">
                Enter transaction ID or user email to view details
              </div>
            </div>
          )}

          {/* Statistics Tab */}
          {activeTab === 'stats' && (
            <div>
              <h2 className="text-xl font-semibold text-gray-900 mb-4">System Statistics</h2>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div className="card">
                  <h3 className="text-lg font-semibold text-gray-900 mb-4">User Stats</h3>
                  <div className="space-y-3">
                    <div className="flex justify-between">
                      <span className="text-gray-600">Total Registered</span>
                      <span className="font-semibold">{stats?.total_users || '-'}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-gray-600">Active (KYC Verified)</span>
                      <span className="font-semibold text-green-600">{stats?.active_users || '-'}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-gray-600">Pending KYC</span>
                      <span className="font-semibold text-orange-600">{stats?.pending_kyc || 0}</span>
                    </div>
                  </div>
                </div>

                <div className="card">
                  <h3 className="text-lg font-semibold text-gray-900 mb-4">Transaction Stats</h3>
                  <div className="space-y-3">
                    <div className="flex justify-between">
                      <span className="text-gray-600">Total Transactions</span>
                      <span className="font-semibold">{stats?.total_transactions || '-'}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-gray-600">Total Wallets</span>
                      <span className="font-semibold">{stats?.total_wallets || '-'}</span>
                    </div>
                  </div>
                </div>

                <div className="card md:col-span-2">
                  <h3 className="text-lg font-semibold text-gray-900 mb-4">Quick Actions</h3>
                  <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                    <button
                      onClick={() => navigate('/kyc')}
                      className="btn-primary py-3"
                    >
                      Review KYC Submissions
                    </button>
                    <button
                      onClick={() => setActiveTab('users')}
                      className="btn-secondary py-3"
                    >
                      Search Users
                    </button>
                    <button
                      onClick={() => setActiveTab('transactions')}
                      className="btn-secondary py-3"
                    >
                      Lookup Transaction
                    </button>
                  </div>
                </div>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
