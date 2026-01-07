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

import { useState, useEffect } from 'react';
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
  const [error, setError] = useState<string | null>(null);
  const [selectedKYC, setSelectedKYC] = useState<KYCWithUser | null>(null);
  const [isProcessing, setIsProcessing] = useState(false);
  const [rejectionReason, setRejectionReason] = useState('');
  const [showRejectModal, setShowRejectModal] = useState(false);

  useEffect(() => {
    fetchPendingKYCs();
  }, []);

  const fetchPendingKYCs = async () => {
    try {
      setIsLoading(true);
      const data = await adminApi.listPendingKYCs();
      setPendingKYCs(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load pending KYCs');
    } finally {
      setIsLoading(false);
    }
  };

  const handleVerify = async (kyc: KYCWithUser) => {
    if (!confirm(`Approve KYC for ${kyc.user.full_name}?`)) return;

    setIsProcessing(true);
    setError(null);

    try {
      await adminApi.verifyKYC(kyc.user.id);

      // Audit log: KYC verification
      if (adminUser) {
        logAdminAction.verifyKYC(
          adminUser.id,
          adminUser.full_name,
          kyc.user.id,
          kyc.user.full_name
        );
      }

      await fetchPendingKYCs();
      setSelectedKYC(null);
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

      await fetchPendingKYCs();
      setShowRejectModal(false);
      setSelectedKYC(null);
      setRejectionReason('');
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
        {/* Error Alert */}
        {error && (
          <Alert variant="error" dismissible onDismiss={() => setError(null)}>
            {error}
          </Alert>
        )}

        {/* Stats */}
        <Card>
          <CardTitle className="mb-2">Pending KYC Submissions</CardTitle>
          {isLoading ? (
            <Skeleton className="h-10 w-20" />
          ) : (
            <p className="text-3xl font-bold text-[var(--interactive-primary)]">{pendingKYCs.length}</p>
          )}
        </Card>

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
        ) : (
          <div className="space-y-6">
            {pendingKYCs.map((item) => (
              <Card key={item.user.id}>
                <div className="flex justify-between items-start mb-4">
                  <div>
                    <h3 className="text-lg font-semibold text-[var(--text-primary)]">{item.user.full_name}</h3>
                    <p className="text-sm text-[var(--text-secondary)]">{item.user.email}</p>
                    <p className="text-sm text-[var(--text-secondary)]">{item.user.phone}</p>
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
                    onClick={() => handleVerify(item)}
                    disabled={isProcessing}
                    loading={isProcessing}
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

      {/* Reject Modal */}
      {showRejectModal && selectedKYC && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center p-4 z-50">
          <Card className="max-w-md w-full">
            <CardTitle className="mb-4">
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
