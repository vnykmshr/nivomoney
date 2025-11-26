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
import { useNavigate } from 'react-router-dom';
import { adminApi } from '../lib/adminApi';
import { formatDateShort } from '@nivo/shared';
import type { KYCWithUser } from '@nivo/shared';
import { useAdminAuthStore } from '../stores/adminAuthStore';
import { logAdminAction } from '../lib/auditLogger';

export function AdminKYC() {
  const navigate = useNavigate();
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

  if (isLoading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-gray-500">Loading pending KYC submissions...</div>
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
              <h1 className="text-xl font-bold text-primary-600">KYC Review</h1>
              <span className="px-2 py-1 bg-purple-100 text-purple-800 text-xs rounded-full">
                Admin Panel
              </span>
            </div>
            <button onClick={() => navigate('/')} className="btn-secondary">
              Back to Dashboard
            </button>
          </div>
        </div>
      </nav>

      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Error Alert */}
        {error && (
          <div className="mb-6 p-4 bg-red-100 text-red-800 rounded-lg flex justify-between items-center">
            <span>{error}</span>
            <button onClick={() => setError(null)} className="text-red-600 hover:text-red-800">
              ✕
            </button>
          </div>
        )}

        {/* Stats */}
        <div className="mb-6 card">
          <h2 className="text-lg font-semibold text-gray-900 mb-2">Pending KYC Submissions</h2>
          <p className="text-3xl font-bold text-primary-600">{pendingKYCs.length}</p>
        </div>

        {/* KYC List */}
        {pendingKYCs.length === 0 ? (
          <div className="card text-center py-12">
            <div className="text-6xl mb-4">✨</div>
            <h3 className="text-lg font-medium text-gray-900 mb-2">All Caught Up!</h3>
            <p className="text-gray-500">No pending KYC submissions</p>
          </div>
        ) : (
          <div className="space-y-6">
            {pendingKYCs.map((item) => (
              <div key={item.user.id} className="card">
                <div className="flex justify-between items-start mb-4">
                  <div>
                    <h3 className="text-lg font-semibold text-gray-900">{item.user.full_name}</h3>
                    <p className="text-sm text-gray-600">{item.user.email}</p>
                    <p className="text-sm text-gray-600">{item.user.phone}</p>
                  </div>
                  <span className="px-3 py-1 bg-yellow-100 text-yellow-800 text-sm rounded-full">
                    Pending Review
                  </span>
                </div>

                {/* KYC Details */}
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4 p-4 bg-gray-50 rounded-lg">
                  <div>
                    <label className="text-xs font-medium text-gray-600 block mb-1">PAN Card</label>
                    <p className="text-sm font-mono font-semibold">{item.kyc.pan}</p>
                  </div>
                  <div>
                    <label className="text-xs font-medium text-gray-600 block mb-1">Date of Birth</label>
                    <p className="text-sm">{item.kyc.date_of_birth ? formatDateShort(item.kyc.date_of_birth) : 'N/A'}</p>
                  </div>
                  <div className="md:col-span-2">
                    <label className="text-xs font-medium text-gray-600 block mb-1">Address</label>
                    {item.kyc.address ? (
                      <p className="text-sm">
                        {item.kyc.address.street}, {item.kyc.address.city}, {item.kyc.address.state} - {item.kyc.address.pin}, {item.kyc.address.country}
                      </p>
                    ) : (
                      <p className="text-sm text-gray-500">No address provided</p>
                    )}
                  </div>
                  <div>
                    <label className="text-xs font-medium text-gray-600 block mb-1">Submitted</label>
                    <p className="text-sm">{formatDateShort(item.kyc.created_at)}</p>
                  </div>
                </div>

                {/* Action Buttons */}
                <div className="flex gap-3">
                  <button
                    onClick={() => handleVerify(item)}
                    disabled={isProcessing}
                    className="flex-1 btn-primary bg-green-600 hover:bg-green-700"
                  >
                    {isProcessing ? 'Processing...' : 'Approve KYC'}
                  </button>
                  <button
                    onClick={() => openRejectModal(item)}
                    disabled={isProcessing}
                    className="flex-1 btn-primary bg-red-600 hover:bg-red-700"
                  >
                    Reject KYC
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}
      </main>

      {/* Reject Modal */}
      {showRejectModal && selectedKYC && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
          <div className="bg-white rounded-lg max-w-md w-full p-6">
            <h3 className="text-lg font-semibold text-gray-900 mb-4">
              Reject KYC for {selectedKYC.user.full_name}
            </h3>

            <div className="mb-4">
              <label htmlFor="rejection-reason" className="block text-sm font-medium text-gray-700 mb-2">
                Rejection Reason <span className="text-red-500">*</span>
              </label>
              <textarea
                id="rejection-reason"
                value={rejectionReason}
                onChange={(e) => setRejectionReason(e.target.value)}
                className="input-field"
                rows={4}
                placeholder="Please provide a clear reason for rejection (minimum 10 characters)"
                disabled={isProcessing}
              />
              <p className="text-sm text-gray-500 mt-1">
                The user will see this reason. Be clear and professional.
              </p>
            </div>

            <div className="flex gap-3">
              <button
                onClick={() => {
                  setShowRejectModal(false);
                  setSelectedKYC(null);
                  setRejectionReason('');
                  setError(null);
                }}
                disabled={isProcessing}
                className="flex-1 btn-secondary"
              >
                Cancel
              </button>
              <button
                onClick={handleReject}
                disabled={isProcessing || !rejectionReason.trim()}
                className="flex-1 btn-primary bg-red-600 hover:bg-red-700"
              >
                {isProcessing ? 'Rejecting...' : 'Confirm Rejection'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
