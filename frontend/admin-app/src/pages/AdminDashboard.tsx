import { useState, useEffect, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { AdminLayout } from '../components';
import { adminApi } from '../lib/adminApi';
import {
  Card,
  CardTitle,
  Button,
  Input,
  Alert,
  Badge,
  Skeleton,
} from '@nivo/shared';
import { cn, getStatusVariant } from '@nivo/shared';
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

type Tab = 'notifications' | 'users' | 'stats';

export function AdminDashboard() {
  const navigate = useNavigate();
  const [activeTab, setActiveTab] = useState<Tab>('notifications');
  const [stats, setStats] = useState<AdminStats | null>(null);
  const [notifications, setNotifications] = useState<NotificationItem[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [lastUpdated, setLastUpdated] = useState<Date | null>(null);
  const [error, setError] = useState<string | null>(null);

  // User search
  const [searchQuery, setSearchQuery] = useState('');
  const [searchResults, setSearchResults] = useState<User[]>([]);
  const [isSearching, setIsSearching] = useState(false);

  const loadDashboardData = useCallback(async (isRefresh = false) => {
    try {
      if (isRefresh) {
        setIsRefreshing(true);
      } else {
        setIsLoading(true);
      }

      const [statsData, pendingKYCs] = await Promise.all([
        adminApi.getAdminStats(),
        adminApi.listPendingKYCs(),
      ]);

      const kycNotifications: NotificationItem[] = pendingKYCs.map(item => ({
        id: item.user.id,
        type: 'kyc_review' as const,
        title: `KYC Review: ${item.user.full_name}`,
        description: `Submitted on ${new Date(item.kyc.created_at).toLocaleDateString()}`,
        priority: 'high' as const,
        link: '/kyc',
        createdAt: item.kyc.created_at,
      }));

      setNotifications(kycNotifications);
      setStats(statsData);
      setLastUpdated(new Date());
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load dashboard data');
    } finally {
      setIsLoading(false);
      setIsRefreshing(false);
    }
  }, []);

  useEffect(() => {
    loadDashboardData();
  }, [loadDashboardData]);

  const handleRefresh = () => {
    loadDashboardData(true);
  };

  const handleSearch = async () => {
    if (!searchQuery.trim()) {
      setSearchResults([]);
      return;
    }

    if (searchQuery.trim().length < 2) {
      setError('Please enter at least 2 characters to search');
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

  const getPriorityVariant = (priority: string): 'error' | 'warning' | 'info' => {
    switch (priority) {
      case 'high': return 'error';
      case 'medium': return 'warning';
      default: return 'info';
    }
  };

  const getNotificationIcon = (type: string) => {
    switch (type) {
      case 'kyc_review':
        return (
          <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
          </svg>
        );
      case 'large_transaction':
        return (
          <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
        );
      default:
        return (
          <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
          </svg>
        );
    }
  };

  const statsCards = [
    {
      label: 'Total Users',
      value: stats?.total_users,
      color: 'text-[var(--text-primary)]',
      bgColor: 'bg-[var(--surface-secondary)]',
      icon: (
        <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
        </svg>
      ),
    },
    {
      label: 'Active Users',
      value: stats?.active_users,
      color: 'text-[var(--color-success-600)]',
      bgColor: 'bg-[var(--color-success-50)]',
      icon: (
        <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
      ),
    },
    {
      label: 'Pending KYC',
      value: stats?.pending_kyc ?? 0,
      color: 'text-[var(--color-primary-600)]',
      bgColor: 'bg-[var(--color-primary-50)]',
      icon: (
        <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
        </svg>
      ),
      action: () => navigate('/kyc'),
    },
    {
      label: 'Total Wallets',
      value: stats?.total_wallets,
      color: 'text-[var(--color-primary-500)]',
      bgColor: 'bg-[var(--color-primary-50)]',
      icon: (
        <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 10h18M7 15h1m4 0h1m-7 4h12a3 3 0 003-3V8a3 3 0 00-3-3H6a3 3 0 00-3 3v8a3 3 0 003 3z" />
        </svg>
      ),
    },
    {
      label: 'Transactions',
      value: stats?.total_transactions,
      color: 'text-[var(--color-accent-600)]',
      bgColor: 'bg-[var(--color-accent-50)]',
      icon: (
        <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 16V4m0 0L3 8m4-4l4 4m6 0v12m0 0l4-4m-4 4l-4-4" />
        </svg>
      ),
      action: () => navigate('/transactions'),
    },
  ];

  const tabs = [
    { id: 'notifications' as Tab, label: 'Notifications', badge: notifications.length },
    { id: 'users' as Tab, label: 'User Management' },
    { id: 'stats' as Tab, label: 'Statistics' },
  ];

  return (
    <AdminLayout title="Dashboard">
      <div className="space-y-6">
        {/* Error Alert */}
        {error && (
          <Alert variant="error" onDismiss={() => setError(null)}>
            {error}
          </Alert>
        )}

        {/* Header with Refresh */}
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold text-[var(--text-primary)]">Dashboard</h1>
            {lastUpdated && (
              <p className="text-sm text-[var(--text-muted)]">
                Last updated: {lastUpdated.toLocaleTimeString()}
              </p>
            )}
          </div>
          <Button
            variant="secondary"
            onClick={handleRefresh}
            loading={isRefreshing}
            disabled={isLoading}
          >
            <svg className="w-4 h-4 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
            </svg>
            Refresh
          </Button>
        </div>

        {/* Quick Actions */}
        <Card className="bg-[var(--surface-brand-subtle)]">
          <div className="flex flex-wrap items-center gap-3">
            <span className="text-sm font-medium text-[var(--text-primary)]">Quick Actions:</span>
            <Button size="sm" onClick={() => navigate('/kyc')}>
              Review KYC {(stats?.pending_kyc ?? 0) > 0 && (
                <Badge variant="error" className="ml-2">{stats?.pending_kyc}</Badge>
              )}
            </Button>
            <Button size="sm" variant="secondary" onClick={() => navigate('/users')}>
              Search Users
            </Button>
            <Button size="sm" variant="secondary" onClick={() => navigate('/transactions')}>
              Search Transactions
            </Button>
          </div>
        </Card>

        {/* Stats Cards */}
        <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-5 gap-4">
          {isLoading ? (
            Array.from({ length: 5 }).map((_, i) => (
              <Card key={i}>
                <Skeleton className="h-4 w-20 mb-2" />
                <Skeleton className="h-8 w-16" />
              </Card>
            ))
          ) : (
            statsCards.map(stat => (
              <Card
                key={stat.label}
                className={cn(
                  stat.action && 'cursor-pointer hover:shadow-md transition-shadow'
                )}
                onClick={stat.action}
              >
                <div className="flex items-start justify-between mb-2">
                  <p className="text-sm text-[var(--text-muted)]">{stat.label}</p>
                  <div className={cn('p-1.5 rounded-lg', stat.bgColor, stat.color)}>
                    {stat.icon}
                  </div>
                </div>
                <p className={cn('text-3xl font-bold', stat.color)}>
                  {stat.value ?? '-'}
                </p>
              </Card>
            ))
          )}
        </div>

        {/* Navigation Tabs */}
        <div className="border-b border-[var(--border-subtle)]">
          <nav className="flex gap-8">
            {tabs.map(tab => (
              <button
                key={tab.id}
                onClick={() => tab.id === 'users' || tab.id === 'notifications' || tab.id === 'stats' ? setActiveTab(tab.id) : navigate('/transactions')}
                className={cn(
                  'py-4 px-1 border-b-2 font-medium text-sm transition-colors',
                  activeTab === tab.id
                    ? 'border-[var(--interactive-primary)] text-[var(--interactive-primary)]'
                    : 'border-transparent text-[var(--text-muted)] hover:text-[var(--text-primary)] hover:border-[var(--border-default)]'
                )}
              >
                {tab.label}
                {tab.badge && tab.badge > 0 && (
                  <Badge variant="error" className="ml-2">
                    {tab.badge}
                  </Badge>
                )}
              </button>
            ))}
            <button
              onClick={() => navigate('/transactions')}
              className="py-4 px-1 border-b-2 border-transparent text-[var(--text-muted)] hover:text-[var(--text-primary)] hover:border-[var(--border-default)] font-medium text-sm transition-colors"
            >
              Transactions
            </button>
          </nav>
        </div>

        {/* Tab Content */}
        {activeTab === 'notifications' && (
          <div>
            <h2 className="text-xl font-semibold text-[var(--text-primary)] mb-4">Pending Actions</h2>

            {isLoading ? (
              <div className="space-y-4">
                {[1, 2, 3].map(i => (
                  <Card key={i}>
                    <div className="flex gap-4">
                      <Skeleton className="w-12 h-12 rounded-lg" />
                      <div className="flex-1">
                        <Skeleton className="h-5 w-48 mb-2" />
                        <Skeleton className="h-4 w-32" />
                      </div>
                    </div>
                  </Card>
                ))}
              </div>
            ) : notifications.length === 0 ? (
              <Card className="text-center py-12">
                <div className="w-16 h-16 mx-auto mb-4 rounded-full bg-[var(--color-success-50)] flex items-center justify-center">
                  <svg className="w-8 h-8 text-[var(--color-success-600)]" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                  </svg>
                </div>
                <CardTitle className="mb-2">All Caught Up!</CardTitle>
                <p className="text-[var(--text-muted)]">No pending notifications at the moment.</p>
              </Card>
            ) : (
              <div className="space-y-4">
                {notifications.map(notification => (
                  <Card
                    key={notification.id}
                    className="cursor-pointer hover:shadow-md transition-shadow"
                    onClick={() => navigate(notification.link)}
                  >
                    <div className="flex items-start justify-between gap-4">
                      <div className="flex items-start gap-4">
                        <div className="w-12 h-12 rounded-lg bg-[var(--surface-brand-subtle)] flex items-center justify-center text-[var(--interactive-primary)]">
                          {getNotificationIcon(notification.type)}
                        </div>
                        <div className="flex-1">
                          <h3 className="text-lg font-semibold text-[var(--text-primary)] mb-1">
                            {notification.title}
                          </h3>
                          <p className="text-sm text-[var(--text-secondary)] mb-2">
                            {notification.description}
                          </p>
                          <div className="flex items-center gap-3">
                            <Badge variant={getPriorityVariant(notification.priority)}>
                              {notification.priority.toUpperCase()}
                            </Badge>
                            <span className="text-xs text-[var(--text-muted)]">
                              {new Date(notification.createdAt).toLocaleString()}
                            </span>
                          </div>
                        </div>
                      </div>
                      <Button>Review</Button>
                    </div>
                  </Card>
                ))}
              </div>
            )}
          </div>
        )}

        {activeTab === 'users' && (
          <div>
            <h2 className="text-xl font-semibold text-[var(--text-primary)] mb-4">User Management</h2>

            {/* Search Bar */}
            <Card className="mb-6">
              <div className="flex gap-3">
                <label htmlFor="user-search" className="sr-only">Search users</label>
                <Input
                  id="user-search"
                  value={searchQuery}
                  onChange={e => setSearchQuery(e.target.value)}
                  onKeyDown={e => e.key === 'Enter' && handleSearch()}
                  placeholder="Search by email, phone, or name..."
                  className="flex-1"
                />
                <Button onClick={handleSearch} loading={isSearching}>
                  Search
                </Button>
              </div>
            </Card>

            {/* Search Results */}
            {searchResults.length > 0 ? (
              <div className="space-y-4">
                {searchResults.map(user => (
                  <Card
                    key={user.id}
                    className="cursor-pointer hover:shadow-md transition-shadow"
                    onClick={() => navigate(`/users/${user.id}`)}
                  >
                    <div className="flex justify-between items-start">
                      <div className="flex-1">
                        <h3 className="text-lg font-semibold text-[var(--text-primary)]">
                          {user.full_name}
                        </h3>
                        <p className="text-sm text-[var(--text-secondary)]">{user.email}</p>
                        <p className="text-sm text-[var(--text-secondary)]">{user.phone}</p>
                        <p className="text-xs text-[var(--text-muted)] mt-1 font-mono">
                          ID: {user.id.slice(0, 8)}...
                        </p>
                      </div>
                      <div className="flex flex-col items-end gap-2">
                        <Badge variant={getStatusVariant(user.status)}>
                          {user.status.toUpperCase()}
                        </Badge>
                        <Button size="sm">View Details</Button>
                      </div>
                    </div>
                  </Card>
                ))}
              </div>
            ) : (
              <Card className="text-center py-12 text-[var(--text-muted)]">
                Search for users by email, phone, or name
              </Card>
            )}
          </div>
        )}

        {activeTab === 'stats' && (
          <div>
            <h2 className="text-xl font-semibold text-[var(--text-primary)] mb-4">System Statistics</h2>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <Card>
                <CardTitle className="mb-4">User Stats</CardTitle>
                <div className="space-y-3">
                  <div className="flex justify-between">
                    <span className="text-[var(--text-secondary)]">Total Registered</span>
                    <span className="font-semibold text-[var(--text-primary)]">{stats?.total_users ?? '-'}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-[var(--text-secondary)]">Active (KYC Verified)</span>
                    <span className="font-semibold text-[var(--color-success-600)]">{stats?.active_users ?? '-'}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-[var(--text-secondary)]">Pending KYC</span>
                    <span className="font-semibold text-[var(--color-primary-600)]">{stats?.pending_kyc ?? 0}</span>
                  </div>
                </div>
              </Card>

              <Card>
                <CardTitle className="mb-4">Transaction Stats</CardTitle>
                <div className="space-y-3">
                  <div className="flex justify-between">
                    <span className="text-[var(--text-secondary)]">Total Transactions</span>
                    <span className="font-semibold text-[var(--text-primary)]">{stats?.total_transactions ?? '-'}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-[var(--text-secondary)]">Total Wallets</span>
                    <span className="font-semibold text-[var(--text-primary)]">{stats?.total_wallets ?? '-'}</span>
                  </div>
                </div>
              </Card>
            </div>
          </div>
        )}
      </div>
    </AdminLayout>
  );
}
