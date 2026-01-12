/**
 * User-Admin Dashboard
 * Shows pending verifications with OTP codes
 */

import { useEffect, useState, useCallback } from 'react';
import { Link } from 'react-router-dom';
import { useAuthStore } from '../stores/authStore';
import { useVerificationStore } from '../stores/verificationStore';
import { api, type VerificationStats } from '../lib/api';
import { OTPDisplay } from '../components/OTPDisplay';
import {
  Card,
  Button,
  Badge,
  Alert,
  Spinner,
  Skeleton,
} from '../../../shared/components';

export function Dashboard() {
  const { user, pairedUser, logout } = useAuthStore();
  const {
    verifications,
    isLoading,
    error,
    lastRefresh,
    fetchPendingVerifications,
    refreshVerifications,
    clearError,
  } = useVerificationStore();

  const [stats, setStats] = useState<VerificationStats | null>(null);

  const fetchStats = useCallback(async () => {
    try {
      const data = await api.getVerificationStats();
      setStats(data);
    } catch {
      // Silently fail - stats are supplementary
    }
  }, []);

  useEffect(() => {
    fetchPendingVerifications();
    fetchStats();
  }, [fetchPendingVerifications, fetchStats]);

  const pendingVerifications = verifications.filter(v => v.status === 'pending');

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
          <span
            className="py-3 border-b-2 border-[var(--color-primary-500)] text-sm font-medium text-[var(--color-primary-600)]"
          >
            Pending
            {pendingVerifications.length > 0 && (
              <span className="ml-2 px-2 py-0.5 bg-[var(--color-primary-100)] text-[var(--color-primary-700)] rounded-full text-xs">
                {pendingVerifications.length}
              </span>
            )}
          </span>
          <Link
            to="/history"
            className="py-3 border-b-2 border-transparent text-sm font-medium text-[var(--text-muted)] hover:text-[var(--text-primary)] transition-colors"
          >
            History
          </Link>
        </nav>
      </div>

      {/* Main Content */}
      <main className="max-w-4xl mx-auto px-4 py-6 space-y-6">
        {/* Statistics */}
        {stats && (
          <Card padding="md">
            <div className="grid grid-cols-3 gap-4 text-center">
              <div>
                <p className="text-2xl font-bold text-[var(--text-primary)]">{stats.today}</p>
                <p className="text-sm text-[var(--text-muted)]">Today</p>
              </div>
              <div>
                <p className="text-2xl font-bold text-[var(--text-primary)]">{stats.this_week}</p>
                <p className="text-sm text-[var(--text-muted)]">This Week</p>
              </div>
              <div>
                <p className="text-2xl font-bold text-[var(--text-primary)]">{stats.total}</p>
                <p className="text-sm text-[var(--text-muted)]">All Time</p>
              </div>
            </div>
          </Card>
        )}
        {/* Paired User Info */}
        {pairedUser && (
          <Card padding="md">
            <div className="flex items-center gap-4">
              <div className="w-12 h-12 rounded-full bg-[var(--color-primary-100)] flex items-center justify-center">
                <span className="text-[var(--color-primary-600)] font-bold text-lg">
                  {pairedUser.full_name.charAt(0).toUpperCase()}
                </span>
              </div>
              <div className="flex-1">
                <p className="text-sm text-[var(--text-muted)]">Paired with:</p>
                <p className="font-semibold text-[var(--text-primary)]">{pairedUser.full_name}</p>
                <p className="text-sm text-[var(--text-secondary)]">{pairedUser.phone}</p>
              </div>
              <Badge variant={pairedUser.kyc_status === 'verified' ? 'success' : 'warning'}>
                KYC: {pairedUser.kyc_status}
              </Badge>
            </div>
          </Card>
        )}

        {/* Error Display */}
        {error && (
          <Alert variant="error" onDismiss={clearError}>
            {error}
          </Alert>
        )}

        {/* Section Header */}
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-xl font-semibold text-[var(--text-primary)]">
              Pending Verifications
            </h2>
            <p className="text-sm text-[var(--text-muted)]">
              {lastRefresh > 0 && `Last updated: ${new Date(lastRefresh).toLocaleTimeString()}`}
            </p>
          </div>
          <Button
            variant="secondary"
            size="sm"
            onClick={refreshVerifications}
            disabled={isLoading}
          >
            {isLoading ? <Spinner size="sm" /> : 'Refresh'}
          </Button>
        </div>

        {/* Loading State */}
        {isLoading && verifications.length === 0 && (
          <div className="space-y-4">
            <Skeleton className="h-48" />
            <Skeleton className="h-48" />
          </div>
        )}

        {/* Verifications List */}
        {!isLoading && pendingVerifications.length === 0 ? (
          <Card padding="lg" className="text-center py-12">
            <div className="relative w-20 h-20 mx-auto mb-6">
              <svg
                className="w-20 h-20 text-[var(--color-primary-100)]"
                fill="currentColor"
                viewBox="0 0 24 24"
                aria-hidden="true"
              >
                <path d="M12 1L3 5v6c0 5.55 3.84 10.74 9 12 5.16-1.26 9-6.45 9-12V5l-9-4z" />
              </svg>
              <div className="absolute inset-0 flex items-center justify-center">
                <svg className="w-8 h-8 text-[var(--color-primary-500)]" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                </svg>
              </div>
            </div>
            <h3 className="text-xl font-semibold text-[var(--text-primary)] mb-2">
              All Caught Up!
            </h3>
            <p className="text-[var(--text-muted)] max-w-sm mx-auto mb-6">
              {pairedUser
                ? `When ${pairedUser.full_name} initiates a transaction that needs verification, the code will appear here automatically.`
                : 'When your paired user initiates a transaction that requires verification, the OTP code will appear here.'}
            </p>
            <Button variant="secondary" onClick={refreshVerifications} disabled={isLoading}>
              {isLoading ? <Spinner size="sm" /> : (
                <>
                  <svg className="w-4 h-4 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                  </svg>
                  Check for Updates
                </>
              )}
            </Button>
          </Card>
        ) : (
          <div className="space-y-4">
            {pendingVerifications.map(verification => (
              <OTPDisplay key={verification.id} verification={verification} />
            ))}
          </div>
        )}

        {/* Help Section */}
        <Card padding="md" className="bg-[var(--color-primary-50)] border-[var(--color-primary-200)]">
          <div className="flex items-start gap-3">
            <svg className="w-5 h-5 text-[var(--color-primary-600)] mt-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            <div className="text-sm">
              <p className="font-medium text-[var(--color-primary-700)] mb-1">How it works:</p>
              <ol className="list-decimal list-inside space-y-1 text-[var(--color-primary-600)]">
                <li>Your paired user initiates a transaction</li>
                <li>A verification code appears here</li>
                <li>Share the code with them to complete the transaction</li>
                <li>Codes expire after 5 minutes</li>
              </ol>
            </div>
          </div>
        </Card>
      </main>
    </div>
  );
}
