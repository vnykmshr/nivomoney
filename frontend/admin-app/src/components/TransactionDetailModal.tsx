/**
 * Transaction Detail Modal
 * Displays comprehensive transaction information for admin investigation
 */

import { useState, useEffect } from 'react';
import { adminApi } from '../lib/adminApi';
import type { Transaction } from '@nivo/shared';

interface TransactionDetailModalProps {
  transactionId: string;
  onClose: () => void;
}

export function TransactionDetailModal({ transactionId, onClose }: TransactionDetailModalProps) {
  const [transaction, setTransaction] = useState<Transaction | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    loadTransactionDetails();
  }, [transactionId]);

  const loadTransactionDetails = async () => {
    try {
      setIsLoading(true);
      setError(null);
      const data = await adminApi.getTransactionDetails(transactionId);
      setTransaction(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load transaction details');
    } finally {
      setIsLoading(false);
    }
  };

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const formatAmount = (amount: number, currency: string) => {
    return `${currency} ${(amount / 100).toFixed(2)}`;
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString('en-IN', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    });
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'completed': return 'bg-green-100 text-green-800';
      case 'pending': return 'bg-yellow-100 text-yellow-800';
      case 'failed': return 'bg-red-100 text-red-800';
      case 'processing': return 'bg-blue-100 text-blue-800';
      case 'reversed': return 'bg-purple-100 text-purple-800';
      case 'cancelled': return 'bg-gray-100 text-gray-800';
      default: return 'bg-gray-100 text-gray-800';
    }
  };

  const getTypeColor = (type: string) => {
    switch (type) {
      case 'deposit': return 'bg-green-100 text-green-800';
      case 'withdrawal': return 'bg-orange-100 text-orange-800';
      case 'transfer': return 'bg-blue-100 text-blue-800';
      case 'reversal': return 'bg-purple-100 text-purple-800';
      case 'fee': return 'bg-red-100 text-red-800';
      case 'refund': return 'bg-teal-100 text-teal-800';
      default: return 'bg-gray-100 text-gray-800';
    }
  };

  // Handle backdrop click
  const handleBackdropClick = (e: React.MouseEvent) => {
    if (e.target === e.currentTarget) {
      onClose();
    }
  };

  if (isLoading) {
    return (
      <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
        <div className="bg-white rounded-lg p-8">
          <div className="text-center">
            <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600 mx-auto"></div>
            <p className="mt-4 text-gray-600">Loading transaction details...</p>
          </div>
        </div>
      </div>
    );
  }

  if (error || !transaction) {
    return (
      <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50" onClick={handleBackdropClick}>
        <div className="bg-white rounded-lg p-8 max-w-md mx-4">
          <div className="text-center">
            <div className="text-red-500 text-5xl mb-4">⚠️</div>
            <h3 className="text-lg font-semibold text-gray-900 mb-2">Error Loading Transaction</h3>
            <p className="text-gray-600 mb-6">{error || 'Transaction not found'}</p>
            <button onClick={onClose} className="btn-primary">
              Close
            </button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4" onClick={handleBackdropClick}>
      <div className="bg-white rounded-lg max-w-3xl w-full max-h-[90vh] overflow-y-auto">
        {/* Header */}
        <div className="sticky top-0 bg-white border-b border-gray-200 px-6 py-4 flex justify-between items-center">
          <div>
            <h2 className="text-xl font-bold text-gray-900">Transaction Details</h2>
            <p className="text-sm text-gray-500 mt-1">Complete transaction information</p>
          </div>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-gray-600 text-2xl leading-none"
          >
            ×
          </button>
        </div>

        {/* Content */}
        <div className="p-6 space-y-6">
          {/* Transaction ID */}
          <div className="bg-gray-50 rounded-lg p-4">
            <label className="block text-sm font-medium text-gray-600 mb-2">Transaction ID</label>
            <div className="flex items-center justify-between">
              <code className="text-sm font-mono text-gray-900 break-all">{transaction.id}</code>
              <button
                onClick={() => copyToClipboard(transaction.id)}
                className="ml-3 px-3 py-1 text-sm bg-white border border-gray-300 rounded hover:bg-gray-50"
              >
                {copied ? 'Copied!' : 'Copy'}
              </button>
            </div>
          </div>

          {/* Status and Type */}
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-600 mb-2">Type</label>
              <span className={`inline-block px-4 py-2 rounded-full text-sm font-semibold ${getTypeColor(transaction.type)}`}>
                {transaction.type.charAt(0).toUpperCase() + transaction.type.slice(1)}
              </span>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-600 mb-2">Status</label>
              <span className={`inline-block px-4 py-2 rounded-full text-sm font-semibold ${getStatusColor(transaction.status)}`}>
                {transaction.status.charAt(0).toUpperCase() + transaction.status.slice(1)}
              </span>
            </div>
          </div>

          {/* Amount */}
          <div>
            <label className="block text-sm font-medium text-gray-600 mb-2">Amount</label>
            <p className="text-3xl font-bold text-gray-900">
              {formatAmount(transaction.amount, transaction.currency)}
            </p>
          </div>

          {/* Description */}
          <div>
            <label className="block text-sm font-medium text-gray-600 mb-2">Description</label>
            <p className="text-gray-900">{transaction.description}</p>
          </div>

          {/* Reference (if exists) */}
          {transaction.reference && (
            <div>
              <label className="block text-sm font-medium text-gray-600 mb-2">Reference</label>
              <p className="text-gray-900 font-mono text-sm">{transaction.reference}</p>
            </div>
          )}

          {/* Wallet IDs */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {transaction.source_wallet_id && (
              <div className="bg-orange-50 rounded-lg p-4">
                <label className="block text-sm font-medium text-orange-800 mb-2">Source Wallet</label>
                <code className="text-sm font-mono text-orange-900 block break-all">{transaction.source_wallet_id}</code>
              </div>
            )}
            {transaction.destination_wallet_id && (
              <div className="bg-green-50 rounded-lg p-4">
                <label className="block text-sm font-medium text-green-800 mb-2">Destination Wallet</label>
                <code className="text-sm font-mono text-green-900 block break-all">{transaction.destination_wallet_id}</code>
              </div>
            )}
          </div>

          {/* Timestamps */}
          <div className="border-t pt-4">
            <h3 className="text-sm font-semibold text-gray-900 mb-3">Timestamps</h3>
            <div className="space-y-2 text-sm">
              <div className="flex justify-between">
                <span className="text-gray-600">Created:</span>
                <span className="text-gray-900 font-mono">{formatDate(transaction.created_at)}</span>
              </div>
              {transaction.processed_at && (
                <div className="flex justify-between">
                  <span className="text-gray-600">Processed:</span>
                  <span className="text-gray-900 font-mono">{formatDate(transaction.processed_at)}</span>
                </div>
              )}
              {transaction.completed_at && (
                <div className="flex justify-between">
                  <span className="text-gray-600">Completed:</span>
                  <span className="text-gray-900 font-mono">{formatDate(transaction.completed_at)}</span>
                </div>
              )}
              <div className="flex justify-between">
                <span className="text-gray-600">Last Updated:</span>
                <span className="text-gray-900 font-mono">{formatDate(transaction.updated_at)}</span>
              </div>
            </div>
          </div>

          {/* Ledger Entry ID */}
          {transaction.ledger_entry_id && (
            <div>
              <label className="block text-sm font-medium text-gray-600 mb-2">Ledger Entry ID</label>
              <code className="text-sm font-mono text-gray-900 bg-gray-50 px-3 py-2 rounded block break-all">
                {transaction.ledger_entry_id}
              </code>
            </div>
          )}

          {/* Parent Transaction (for reversals/refunds) */}
          {transaction.parent_transaction_id && (
            <div>
              <label className="block text-sm font-medium text-gray-600 mb-2">Parent Transaction ID</label>
              <code className="text-sm font-mono text-gray-900 bg-gray-50 px-3 py-2 rounded block break-all">
                {transaction.parent_transaction_id}
              </code>
            </div>
          )}

          {/* Failure Reason */}
          {transaction.failure_reason && (
            <div className="bg-red-50 rounded-lg p-4">
              <label className="block text-sm font-medium text-red-800 mb-2">Failure Reason</label>
              <p className="text-red-900">{transaction.failure_reason}</p>
            </div>
          )}

          {/* Metadata */}
          {transaction.metadata && Object.keys(transaction.metadata).length > 0 && (
            <div>
              <label className="block text-sm font-medium text-gray-600 mb-2">Metadata</label>
              <div className="bg-gray-50 rounded-lg p-4 space-y-2">
                {Object.entries(transaction.metadata).map(([key, value]) => (
                  <div key={key} className="flex justify-between text-sm">
                    <span className="text-gray-600 font-medium">{key}:</span>
                    <span className="text-gray-900">{value}</span>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>

        {/* Footer Actions */}
        <div className="sticky bottom-0 bg-gray-50 border-t border-gray-200 px-6 py-4 flex justify-end space-x-3">
          <button onClick={onClose} className="btn-secondary">
            Close
          </button>
        </div>
      </div>
    </div>
  );
}
