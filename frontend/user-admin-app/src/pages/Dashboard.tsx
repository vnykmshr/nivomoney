/**
 * User-Admin Dashboard
 * Shows pending verifications with OTP codes
 */

import { useEffect } from 'react';
import { useAuthStore } from '../stores/authStore';
import { useVerificationStore } from '../stores/verificationStore';
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

  useEffect(() => {
    fetchPendingVerifications();
  }, [fetchPendingVerifications]);

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

      {/* Main Content */}
      <main className="max-w-4xl mx-auto px-4 py-6 space-y-6">
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
            <svg
              className="mx-auto h-16 w-16 text-[var(--text-muted)] mb-4"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={1.5}
                d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z"
              />
            </svg>
            <h3 className="text-lg font-medium text-[var(--text-primary)] mb-2">
              No Pending Verifications
            </h3>
            <p className="text-[var(--text-muted)] max-w-sm mx-auto">
              When your paired user initiates a transaction that requires verification,
              the OTP code will appear here.
            </p>
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
