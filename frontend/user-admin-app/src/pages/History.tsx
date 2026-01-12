/**
 * Verification History Page
 * Shows completed verifications with filtering
 */

import { useState, useEffect, useCallback, useMemo } from 'react';
import { Link } from 'react-router-dom';
import { useAuthStore } from '../stores/authStore';
import { api, type Verification } from '../lib/api';
import {
  Card,
  Button,
  Badge,
  Alert,
  Skeleton,
} from '../../../shared/components';

type DateRange = 'today' | 'week' | 'month' | 'all';

export function History() {
  const { user, logout } = useAuthStore();
  const [verifications, setVerifications] = useState<Verification[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [dateRange, setDateRange] = useState<DateRange>('all');

  const fetchHistory = useCallback(async () => {
    try {
      setIsLoading(true);
      setError(null);
      const data = await api.getCompletedVerifications();
      setVerifications(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load history');
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchHistory();
  }, [fetchHistory]);

  // Filter verifications by date range (memoized for performance)
  const filteredVerifications = useMemo(() => {
    return verifications.filter(v => {
      if (dateRange === 'all') return true;

      const verifiedDate = v.verified_at ? new Date(v.verified_at) : new Date(v.created_at);
      const now = new Date();

      switch (dateRange) {
        case 'today': {
          const todayStart = new Date(now.getFullYear(), now.getMonth(), now.getDate());
          return verifiedDate >= todayStart;
        }
        case 'week': {
          const weekAgo = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000);
          return verifiedDate >= weekAgo;
        }
        case 'month': {
          const monthAgo = new Date(now.getTime() - 30 * 24 * 60 * 60 * 1000);
          return verifiedDate >= monthAgo;
        }
        default:
          return true;
      }
    });
  }, [verifications, dateRange]);

  const getOperationLabel = (type: string): string => {
    const labels: Record<string, string> = {
      withdraw: 'Withdrawal',
      transfer: 'Money Transfer',
      add_beneficiary: 'Add Beneficiary',
      password_change: 'Password Change',
      profile_update: 'Profile Update',
    };
    return labels[type] || type.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase());
  };

  const getStatusVariant = (status: string): 'success' | 'error' | 'warning' | 'info' => {
    switch (status) {
      case 'verified': return 'success';
      case 'expired': return 'error';
      case 'cancelled': return 'warning';
      default: return 'info';
    }
  };

  const formatDate = (dateStr: string): string => {
    const date = new Date(dateStr);
    return date.toLocaleDateString('en-IN', {
      day: 'numeric',
      month: 'short',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  return (
    <div className="min-h-screen bg-[var(--surface-page)]">
      {/* Header */}
      <header className="h-16 bg-[var(--surface-card)] border-b border-[var(--border-default)] px-4">
        <div className="max-w-4xl mx-auto h-full flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-full bg-[var(--color-primary-500)] flex items-center justify-center">
              <span className="text-white font-bold">N</span>
            </div>
            <div>
              <div className="flex items-center gap-2">
                <h1 className="font-semibold text-[var(--text-primary)]">Nivo Money</h1>
                <Badge variant="info">Verification Portal</Badge>
              </div>
              <p className="text-sm text-[var(--text-muted)]">
                {user?.full_name}
              </p>
            </div>
          </div>
          <Button variant="secondary" size="sm" onClick={logout}>
            Logout
          </Button>
        </div>
      </header>

      {/* Tab Navigation */}
      <div className="border-b border-[var(--border-default)] bg-[var(--surface-card)]">
        <nav className="max-w-4xl mx-auto px-4 flex gap-6">
          <Link
            to="/"
            className="py-3 border-b-2 border-transparent text-sm font-medium text-[var(--text-muted)] hover:text-[var(--text-primary)] transition-colors"
          >
            Pending
          </Link>
          <span
            className="py-3 border-b-2 border-[var(--color-primary-500)] text-sm font-medium text-[var(--color-primary-600)]"
          >
            History
          </span>
        </nav>
      </div>

      {/* Main Content */}
      <main className="max-w-4xl mx-auto px-4 py-6 space-y-6">
        {/* Error Display */}
        {error && (
          <Alert variant="error" onDismiss={() => setError(null)}>
            {error}
          </Alert>
        )}

        {/* Filter Card */}
        <Card padding="md">
          <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
            <h2 className="font-semibold text-[var(--text-primary)]">
              Verification History
            </h2>
            <div className="flex items-center gap-2">
              <label htmlFor="date-range" className="sr-only">Filter by date</label>
              <select
                id="date-range"
                value={dateRange}
                onChange={(e) => setDateRange(e.target.value as DateRange)}
                className="px-3 py-2 rounded-lg border bg-[var(--surface-input)] border-[var(--border-default)] text-[var(--text-primary)] text-sm focus:outline-none focus:ring-2 focus:ring-[var(--interactive-primary)]"
              >
                <option value="today">Today</option>
                <option value="week">This Week</option>
                <option value="month">This Month</option>
                <option value="all">All Time</option>
              </select>
              <Button variant="secondary" size="sm" onClick={fetchHistory} disabled={isLoading}>
                Refresh
              </Button>
            </div>
          </div>
        </Card>

        {/* Loading State */}
        {isLoading && (
          <div className="space-y-4">
            <Skeleton className="h-24" />
            <Skeleton className="h-24" />
            <Skeleton className="h-24" />
          </div>
        )}

        {/* Empty State */}
        {!isLoading && filteredVerifications.length === 0 && (
          <Card padding="lg" className="text-center py-12">
            <svg
              className="mx-auto h-16 w-16 text-[var(--text-muted)] mb-4"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              aria-hidden="true"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={1.5}
                d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
              />
            </svg>
            <h3 className="text-lg font-medium text-[var(--text-primary)] mb-2">
              No Verification History
            </h3>
            <p className="text-[var(--text-muted)] max-w-sm mx-auto mb-4">
              {dateRange === 'all'
                ? 'Completed verifications will appear here once you process your first one.'
                : `No verifications found for the selected time period. Try expanding your date range.`}
            </p>
            {dateRange !== 'all' && (
              <Button variant="secondary" size="sm" onClick={() => setDateRange('all')}>
                View All Time
              </Button>
            )}
          </Card>
        )}

        {/* History List */}
        {!isLoading && filteredVerifications.length > 0 && (
          <div className="space-y-3">
            <p className="text-sm text-[var(--text-muted)]">
              Showing {filteredVerifications.length} verification{filteredVerifications.length !== 1 ? 's' : ''}
            </p>
            {filteredVerifications.map(v => (
              <Card key={v.id} padding="md">
                <div className="flex items-center justify-between">
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 mb-1">
                      <p className="font-medium text-[var(--text-primary)]">
                        {getOperationLabel(v.operation_type)}
                      </p>
                      <Badge variant={getStatusVariant(v.status)}>
                        {v.status}
                      </Badge>
                    </div>
                    {v.metadata && (
                      <p className="text-sm text-[var(--text-secondary)] truncate">
                        {v.metadata.amount && `Amount: ${v.metadata.amount}`}
                        {v.metadata.beneficiary && ` • To: ${v.metadata.beneficiary}`}
                      </p>
                    )}
                  </div>
                  <div className="text-right ml-4 flex-shrink-0">
                    <p className="text-sm text-[var(--text-muted)]">
                      {formatDate(v.verified_at || v.created_at)}
                    </p>
                  </div>
                </div>
              </Card>
            ))}
          </div>
        )}
      </main>
    </div>
  );
}
