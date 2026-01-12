/**
 * Admin KYC Review Panel
 * Migrated from user-app to admin-app
 *
 * Admin Workflow Pattern:
 * 1. User submits KYC → Notification sent to admins
 * 2. Admin sees notification in dashboard
 * 3. Admin clicks notification → comes to this review page
 * 4. Admin approves/rejects → User notified
 *
 * This page can also be accessed directly for batch processing
 * or when notification is missed.
 *
 * See: /ADMIN_WORKFLOW_PATTERN.md
 */

import { useState, useEffect, useCallback } from 'react';
import { AdminLayout } from '../components';
import { adminApi } from '../lib/adminApi';
import { formatDateShort } from '@nivo/shared';
import type { KYCWithUser } from '@nivo/shared';
import { useAdminAuthStore } from '../stores/adminAuthStore';
import { logAdminAction } from '../lib/auditLogger';
import {
  Card,
  CardTitle,
  Button,
  Input,
  Alert,
  Badge,
  Skeleton,
  FormField,
} from '../../../shared/components';
import { cn } from '../../../shared/lib/utils';

export function AdminKYC() {
  const { user: adminUser } = useAdminAuthStore();
  const [pendingKYCs, setPendingKYCs] = useState<KYCWithUser[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [selectedKYC, setSelectedKYC] = useState<KYCWithUser | null>(null);
  const [isProcessing, setIsProcessing] = useState(false);
  const [rejectionReason, setRejectionReason] = useState('');
  const [showRejectModal, setShowRejectModal] = useState(false);
  const [showApproveModal, setShowApproveModal] = useState(false);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);

  // Search and bulk selection
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set());

  const fetchPendingKYCs = useCallback(async (isRefresh = false) => {
    try {
      if (isRefresh) {
        setIsRefreshing(true);
      } else {
        setIsLoading(true);
      }
      const data = await adminApi.listPendingKYCs();
      setPendingKYCs(data);
      setSelectedIds(new Set()); // Clear selection on refresh
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load pending KYCs');
    } finally {
      setIsLoading(false);
      setIsRefreshing(false);
    }
  }, []);

  useEffect(() => {
    fetchPendingKYCs();
  }, [fetchPendingKYCs]);

  // Filter KYCs based on search query
  const filteredKYCs = pendingKYCs.filter(item => {
    if (!searchQuery.trim()) return true;
    const query = searchQuery.toLowerCase();
    return (
      item.user.full_name.toLowerCase().includes(query) ||
      item.user.email.toLowerCase().includes(query) ||
      item.user.phone.includes(query) ||
      item.kyc.pan?.toLowerCase().includes(query)
    );
  });

  // Selection handlers
  const toggleSelection = (userId: string) => {
    const newSelection = new Set(selectedIds);
    if (newSelection.has(userId)) {
      newSelection.delete(userId);
    } else {
      newSelection.add(userId);
    }
    setSelectedIds(newSelection);
  };

  const toggleSelectAll = () => {
    if (selectedIds.size === filteredKYCs.length) {
      setSelectedIds(new Set());
    } else {
      setSelectedIds(new Set(filteredKYCs.map(item => item.user.id)));
    }
  };

  const handleRefresh = () => {
    fetchPendingKYCs(true);
  };

  // Handle escape key to close modals
  useEffect(() => {
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        if (showRejectModal) {
          setShowRejectModal(false);
          setSelectedKYC(null);
          setRejectionReason('');
        }
        if (showApproveModal) {
          setShowApproveModal(false);
          setSelectedKYC(null);
        }
      }
    };
    document.addEventListener('keydown', handleEscape);
    return () => document.removeEventListener('keydown', handleEscape);
  }, [showRejectModal, showApproveModal]);

  const openApproveModal = (kyc: KYCWithUser) => {
    setSelectedKYC(kyc);
    setShowApproveModal(true);
    setError(null);
  };

  const handleVerify = async () => {
    if (!selectedKYC) return;

    setIsProcessing(true);
    setError(null);

    try {
      await adminApi.verifyKYC(selectedKYC.user.id);

      // Audit log: KYC verification
      if (adminUser) {
        logAdminAction.verifyKYC(
          adminUser.id,
          adminUser.full_name,
          selectedKYC.user.id,
          selectedKYC.user.full_name
        );
      }

      const userName = selectedKYC.user.full_name;
      await fetchPendingKYCs();
      setShowApproveModal(false);
      setSelectedKYC(null);
      setSuccessMessage(`KYC approved for ${userName}. User now has full wallet access.`);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to verify KYC');
    } finally {
      setIsProcessing(false);
    }
  };

  const handleReject = async () => {
    if (!selectedKYC || !rejectionReason.trim()) {
      setError('Rejection reason is required');
      return;
    }

    if (rejectionReason.length < 10) {
      setError('Rejection reason must be at least 10 characters');
      return;
    }

    setIsProcessing(true);
    setError(null);

    try {
      await adminApi.rejectKYC(selectedKYC.user.id, rejectionReason);

      // Audit log: KYC rejection
      if (adminUser) {
        logAdminAction.rejectKYC(
          adminUser.id,
          adminUser.full_name,
          selectedKYC.user.id,
          selectedKYC.user.full_name,
          rejectionReason
        );
      }

      const userName = selectedKYC.user.full_name;
      await fetchPendingKYCs();
      setShowRejectModal(false);
      setSelectedKYC(null);
      setRejectionReason('');
      setSuccessMessage(`KYC rejected for ${userName}. The user has been notified.`);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to reject KYC');
    } finally {
      setIsProcessing(false);
    }
  };

  const openRejectModal = (kyc: KYCWithUser) => {
    setSelectedKYC(kyc);
    setShowRejectModal(true);
    setRejectionReason('');
    setError(null);
  };

  return (
    <AdminLayout title="KYC Review">
      <div className="space-y-6">
        {/* Success Alert */}
        {successMessage && (
          <Alert variant="success" onDismiss={() => setSuccessMessage(null)}>
            {successMessage}
          </Alert>
        )}

        {/* Error Alert */}
        {error && (
          <Alert variant="error" onDismiss={() => setError(null)}>
            {error}
          </Alert>
        )}

        {/* Header with Stats and Actions */}
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
          <div>
            <h1 className="text-2xl font-bold text-[var(--text-primary)]">KYC Review</h1>
            <p className="text-sm text-[var(--text-muted)]">
              {pendingKYCs.length} pending submission{pendingKYCs.length !== 1 ? 's' : ''}
            </p>
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

        {/* Search */}
        {pendingKYCs.length > 0 && (
          <Card>
            <div className="flex gap-3">
              <label htmlFor="kyc-search" className="sr-only">Search KYC submissions</label>
              <Input
                id="kyc-search"
                value={searchQuery}
                onChange={e => setSearchQuery(e.target.value)}
                placeholder="Search by name, email, phone, or PAN..."
                className="flex-1"
              />
              {searchQuery && (
                <Button variant="secondary" onClick={() => setSearchQuery('')}>
                  Clear
                </Button>
              )}
            </div>
            {searchQuery && (
              <p className="text-sm text-[var(--text-muted)] mt-2">
                Showing {filteredKYCs.length} of {pendingKYCs.length} submissions
              </p>
            )}
          </Card>
        )}

        {/* Bulk Selection Bar */}
        {selectedIds.size > 0 && (
          <Card className="bg-[var(--surface-brand-subtle)]">
            <div className="flex items-center justify-between">
              <span className="text-sm font-medium text-[var(--text-primary)]">
                {selectedIds.size} selected
              </span>
              <div className="flex gap-2">
                <Button
                  size="sm"
                  onClick={() => setSelectedIds(new Set())}
                  variant="secondary"
                >
                  Clear Selection
                </Button>
              </div>
            </div>
          </Card>
        )}

        {/* KYC List */}
        {isLoading ? (
          <div className="space-y-6">
            {[1, 2, 3].map(i => (
              <Card key={i}>
                <div className="flex justify-between items-start mb-4">
                  <div>
                    <Skeleton className="h-6 w-40 mb-2" />
                    <Skeleton className="h-4 w-48 mb-1" />
                    <Skeleton className="h-4 w-32" />
                  </div>
                  <Skeleton className="h-6 w-28" />
                </div>
                <Skeleton className="h-32 w-full mb-4" />
                <div className="flex gap-3">
                  <Skeleton className="h-10 flex-1" />
                  <Skeleton className="h-10 flex-1" />
                </div>
              </Card>
            ))}
          </div>
        ) : pendingKYCs.length === 0 ? (
          <Card className="text-center py-12">
            <div className="w-16 h-16 mx-auto mb-4 rounded-full bg-[var(--color-success-50)] flex items-center justify-center">
              <svg className="w-8 h-8 text-[var(--color-success-600)]" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
              </svg>
            </div>
            <CardTitle className="mb-2">All Caught Up!</CardTitle>
            <p className="text-[var(--text-muted)]">No pending KYC submissions</p>
          </Card>
        ) : filteredKYCs.length === 0 ? (
          <Card className="text-center py-12">
            <CardTitle className="mb-2">No Matches</CardTitle>
            <p className="text-[var(--text-muted)]">No submissions match your search criteria</p>
            <Button variant="secondary" className="mt-4" onClick={() => setSearchQuery('')}>
              Clear Search
            </Button>
          </Card>
        ) : (
          <div className="space-y-6">
            {/* Select All */}
            {filteredKYCs.length > 1 && (
              <div className="flex items-center gap-3">
                <label className="flex items-center gap-2 cursor-pointer">
                  <input
                    type="checkbox"
                    checked={selectedIds.size === filteredKYCs.length && filteredKYCs.length > 0}
                    onChange={toggleSelectAll}
                    className="w-4 h-4 rounded border-[var(--border-default)] text-[var(--interactive-primary)] focus:ring-[var(--interactive-primary)]"
                  />
                  <span className="text-sm text-[var(--text-secondary)]">Select all</span>
                </label>
              </div>
            )}

            {filteredKYCs.map((item) => (
              <Card key={item.user.id} className={cn(
                selectedIds.has(item.user.id) && 'ring-2 ring-[var(--interactive-primary)]'
              )}>
                <div className="flex justify-between items-start mb-4">
                  <div className="flex items-start gap-3">
                    <input
                      type="checkbox"
                      checked={selectedIds.has(item.user.id)}
                      onChange={() => toggleSelection(item.user.id)}
                      className="mt-1 w-4 h-4 rounded border-[var(--border-default)] text-[var(--interactive-primary)] focus:ring-[var(--interactive-primary)]"
                    />
                    <div>
                      <h3 className="text-lg font-semibold text-[var(--text-primary)]">{item.user.full_name}</h3>
                      <p className="text-sm text-[var(--text-secondary)]">{item.user.email}</p>
                      <p className="text-sm text-[var(--text-secondary)]">{item.user.phone}</p>
                    </div>
                  </div>
                  <Badge variant="warning">Pending Review</Badge>
                </div>

                {/* KYC Details */}
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4 p-4 bg-[var(--surface-secondary)] rounded-lg">
                  <div>
                    <label className="text-xs font-medium text-[var(--text-muted)] block mb-1">PAN Card</label>
                    <p className="text-sm font-mono font-semibold text-[var(--text-primary)]">{item.kyc.pan}</p>
                  </div>
                  <div>
                    <label className="text-xs font-medium text-[var(--text-muted)] block mb-1">Date of Birth</label>
                    <p className="text-sm text-[var(--text-primary)]">{item.kyc.date_of_birth ? formatDateShort(item.kyc.date_of_birth) : 'N/A'}</p>
                  </div>
                  <div className="md:col-span-2">
                    <label className="text-xs font-medium text-[var(--text-muted)] block mb-1">Address</label>
                    {item.kyc.address ? (
                      <p className="text-sm text-[var(--text-primary)]">
                        {item.kyc.address.street}, {item.kyc.address.city}, {item.kyc.address.state} - {item.kyc.address.pin}, {item.kyc.address.country}
                      </p>
                    ) : (
                      <p className="text-sm text-[var(--text-muted)]">No address provided</p>
                    )}
                  </div>
                  <div>
                    <label className="text-xs font-medium text-[var(--text-muted)] block mb-1">Submitted</label>
                    <p className="text-sm text-[var(--text-primary)]">{formatDateShort(item.kyc.created_at)}</p>
                  </div>
                </div>

                {/* Action Buttons */}
                <div className="flex gap-3">
                  <Button
                    onClick={() => openApproveModal(item)}
                    disabled={isProcessing}
                    className="flex-1 bg-[var(--color-success-600)] hover:bg-[var(--color-success-700)]"
                  >
                    Approve KYC
                  </Button>
                  <Button
                    onClick={() => openRejectModal(item)}
                    disabled={isProcessing}
                    className="flex-1 bg-[var(--color-error-600)] hover:bg-[var(--color-error-700)]"
                  >
                    Reject KYC
                  </Button>
                </div>
              </Card>
            ))}
          </div>
        )}
      </div>

      {/* Approve Modal */}
      {showApproveModal && selectedKYC && (
        <div
          className="fixed inset-0 bg-[var(--surface-overlay)] flex items-center justify-center p-4 z-50"
          role="dialog"
          aria-modal="true"
          aria-labelledby="approve-modal-title"
        >
          <Card className="max-w-md w-full">
            <CardTitle id="approve-modal-title" className="mb-4">
              Approve KYC for {selectedKYC.user.full_name}
            </CardTitle>

            <p className="text-sm text-[var(--text-secondary)] mb-4">
              You are about to verify the KYC submission for this user.
              This will grant them full access to wallet features.
            </p>

            <div className="bg-[var(--surface-secondary)] rounded-lg p-4 mb-4">
              <div className="grid grid-cols-2 gap-2 text-sm">
                <div>
                  <p className="text-[var(--text-muted)]">Name</p>
                  <p className="font-medium text-[var(--text-primary)]">{selectedKYC.user.full_name}</p>
                </div>
                <div>
                  <p className="text-[var(--text-muted)]">PAN</p>
                  <p className="font-mono font-medium text-[var(--text-primary)]">{selectedKYC.kyc.pan}</p>
                </div>
              </div>
            </div>

            <div className="flex gap-3">
              <Button
                variant="secondary"
                onClick={() => {
                  setShowApproveModal(false);
                  setSelectedKYC(null);
                }}
                disabled={isProcessing}
                className="flex-1"
              >
                Cancel
              </Button>
              <Button
                onClick={handleVerify}
                disabled={isProcessing}
                loading={isProcessing}
                className="flex-1 bg-[var(--color-success-600)] hover:bg-[var(--color-success-700)]"
              >
                Confirm Approval
              </Button>
            </div>
          </Card>
        </div>
      )}

      {/* Reject Modal */}
      {showRejectModal && selectedKYC && (
        <div
          className="fixed inset-0 bg-[var(--surface-overlay)] flex items-center justify-center p-4 z-50"
          role="dialog"
          aria-modal="true"
          aria-labelledby="reject-modal-title"
        >
          <Card className="max-w-md w-full">
            <CardTitle id="reject-modal-title" className="mb-4">
              Reject KYC for {selectedKYC.user.full_name}
            </CardTitle>

            <FormField
              label="Rejection Reason"
              htmlFor="rejection-reason"
              required
              hint="The user will see this reason. Be clear and professional."
            >
              <textarea
                id="rejection-reason"
                value={rejectionReason}
                onChange={(e) => setRejectionReason(e.target.value)}
                className={cn(
                  'w-full px-4 py-3 rounded-lg border transition-colors resize-none',
                  'bg-[var(--surface-input)] border-[var(--border-default)]',
                  'text-[var(--text-primary)] placeholder:text-[var(--text-muted)]',
                  'focus:outline-none focus:ring-2 focus:ring-[var(--interactive-primary)] focus:border-transparent'
                )}
                rows={4}
                placeholder="Please provide a clear reason for rejection (minimum 10 characters)"
                disabled={isProcessing}
              />
            </FormField>

            <div className="flex gap-3 mt-4">
              <Button
                variant="secondary"
                onClick={() => {
                  setShowRejectModal(false);
                  setSelectedKYC(null);
                  setRejectionReason('');
                  setError(null);
                }}
                disabled={isProcessing}
                className="flex-1"
              >
                Cancel
              </Button>
              <Button
                onClick={handleReject}
                disabled={isProcessing || !rejectionReason.trim()}
                loading={isProcessing}
                className="flex-1 bg-[var(--color-error-600)] hover:bg-[var(--color-error-700)]"
              >
                Confirm Rejection
              </Button>
            </div>
          </Card>
        </div>
      )}
    </AdminLayout>
  );
}
