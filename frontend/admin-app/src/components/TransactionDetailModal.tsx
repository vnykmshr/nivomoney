/**
 * Transaction Detail Modal
 * Displays comprehensive transaction information for admin investigation
 */

import { useState, useEffect, useCallback } from 'react';
import { adminApi } from '../lib/adminApi';
import type { Transaction } from '@nivo/shared';
import {
  Card,
  CardTitle,
  Button,
  Badge,
  Skeleton,
} from '../../../shared/components';
import {
  getTransactionStatusVariant,
  getTransactionTypeVariant,
} from '../../../shared/lib';

interface TransactionDetailModalProps {
  transactionId: string;
  onClose: () => void;
}

export function TransactionDetailModal({ transactionId, onClose }: TransactionDetailModalProps) {
  const [transaction, setTransaction] = useState<Transaction | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [copied, setCopied] = useState(false);

  const loadTransactionDetails = useCallback(async () => {
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
  }, [transactionId]);

  useEffect(() => {
    loadTransactionDetails();
  }, [loadTransactionDetails]);

  // Handle escape key
  useEffect(() => {
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        onClose();
      }
    };
    document.addEventListener('keydown', handleEscape);
    return () => document.removeEventListener('keydown', handleEscape);
  }, [onClose]);

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

  // Handle backdrop click
  const handleBackdropClick = (e: React.MouseEvent) => {
    if (e.target === e.currentTarget) {
      onClose();
    }
  };

  if (isLoading) {
    return (
      <div className="fixed inset-0 bg-[var(--surface-overlay)] flex items-center justify-center z-50 p-4">
        <Card className="max-w-3xl w-full p-8">
          <div className="text-center">
            <Skeleton className="h-12 w-12 rounded-full mx-auto mb-4" />
            <Skeleton className="h-5 w-48 mx-auto" />
          </div>
        </Card>
      </div>
    );
  }

  if (error || !transaction) {
    return (
      <div className="fixed inset-0 bg-[var(--surface-overlay)] flex items-center justify-center z-50 p-4" onClick={handleBackdropClick}>
        <Card className="max-w-md w-full text-center py-8">
          <div className="w-16 h-16 mx-auto mb-4 rounded-full bg-[var(--color-error-50)] flex items-center justify-center">
            <svg className="w-8 h-8 text-[var(--color-error-600)]" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
            </svg>
          </div>
          <CardTitle className="mb-2">Error Loading Transaction</CardTitle>
          <p className="text-[var(--text-muted)] mb-6">{error || 'Transaction not found'}</p>
          <Button onClick={onClose}>Close</Button>
        </Card>
      </div>
    );
  }

  return (
    <div className="fixed inset-0 bg-[var(--surface-overlay)] flex items-center justify-center z-50 p-4" onClick={handleBackdropClick}>
      <Card className="max-w-3xl w-full max-h-[90vh] overflow-y-auto p-0">
        {/* Header */}
        <div className="sticky top-0 bg-[var(--surface-card)] border-b border-[var(--border-subtle)] px-6 py-4 flex justify-between items-center">
          <div>
            <h2 className="text-xl font-bold text-[var(--text-primary)]">Transaction Details</h2>
            <p className="text-sm text-[var(--text-muted)] mt-1">Complete transaction information</p>
          </div>
          <button
            onClick={onClose}
            className="text-[var(--text-muted)] hover:text-[var(--text-primary)] text-2xl leading-none p-2"
            aria-label="Close modal"
          >
            &times;
          </button>
        </div>

        {/* Content */}
        <div className="p-6 space-y-6">
          {/* Transaction ID */}
          <div className="bg-[var(--surface-secondary)] rounded-lg p-4">
            <p className="text-sm font-medium text-[var(--text-muted)] mb-2">Transaction ID</p>
            <div className="flex items-center justify-between">
              <code className="text-sm font-mono text-[var(--text-primary)] break-all">{transaction.id}</code>
              <Button
                onClick={() => copyToClipboard(transaction.id)}
                variant="secondary"
                size="sm"
              >
                {copied ? 'Copied!' : 'Copy'}
              </Button>
            </div>
          </div>

          {/* Status and Type */}
          <div className="grid grid-cols-2 gap-4">
            <div>
              <p className="text-sm font-medium text-[var(--text-muted)] mb-2">Type</p>
              <Badge variant={getTransactionTypeVariant(transaction.type)}>
                {transaction.type.charAt(0).toUpperCase() + transaction.type.slice(1)}
              </Badge>
            </div>
            <div>
              <p className="text-sm font-medium text-[var(--text-muted)] mb-2">Status</p>
              <Badge variant={getTransactionStatusVariant(transaction.status)}>
                {transaction.status.charAt(0).toUpperCase() + transaction.status.slice(1)}
              </Badge>
            </div>
          </div>

          {/* Amount */}
          <div>
            <p className="text-sm font-medium text-[var(--text-muted)] mb-2">Amount</p>
            <p className="text-3xl font-bold text-[var(--text-primary)]">
              {formatAmount(transaction.amount, transaction.currency)}
            </p>
          </div>

          {/* Description */}
          <div>
            <p className="text-sm font-medium text-[var(--text-muted)] mb-2">Description</p>
            <p className="text-[var(--text-primary)]">{transaction.description}</p>
          </div>

          {/* Reference (if exists) */}
          {transaction.reference && (
            <div>
              <p className="text-sm font-medium text-[var(--text-muted)] mb-2">Reference</p>
              <p className="text-[var(--text-primary)] font-mono text-sm">{transaction.reference}</p>
            </div>
          )}

          {/* Wallet IDs */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {transaction.source_wallet_id && (
              <div className="bg-[var(--color-warning-50)] rounded-lg p-4">
                <p className="text-sm font-medium text-[var(--color-warning-800)] mb-2">Source Wallet</p>
                <code className="text-sm font-mono text-[var(--color-warning-900)] block break-all">{transaction.source_wallet_id}</code>
              </div>
            )}
            {transaction.destination_wallet_id && (
              <div className="bg-[var(--color-success-50)] rounded-lg p-4">
                <p className="text-sm font-medium text-[var(--color-success-800)] mb-2">Destination Wallet</p>
                <code className="text-sm font-mono text-[var(--color-success-900)] block break-all">{transaction.destination_wallet_id}</code>
              </div>
            )}
          </div>

          {/* Timestamps */}
          <div className="border-t border-[var(--border-subtle)] pt-4">
            <h3 className="text-sm font-semibold text-[var(--text-primary)] mb-3">Timestamps</h3>
            <div className="space-y-2 text-sm">
              <div className="flex justify-between">
                <span className="text-[var(--text-muted)]">Created:</span>
                <span className="text-[var(--text-primary)] font-mono">{formatDate(transaction.created_at)}</span>
              </div>
              {transaction.processed_at && (
                <div className="flex justify-between">
                  <span className="text-[var(--text-muted)]">Processed:</span>
                  <span className="text-[var(--text-primary)] font-mono">{formatDate(transaction.processed_at)}</span>
                </div>
              )}
              {transaction.completed_at && (
                <div className="flex justify-between">
                  <span className="text-[var(--text-muted)]">Completed:</span>
                  <span className="text-[var(--text-primary)] font-mono">{formatDate(transaction.completed_at)}</span>
                </div>
              )}
              <div className="flex justify-between">
                <span className="text-[var(--text-muted)]">Last Updated:</span>
                <span className="text-[var(--text-primary)] font-mono">{formatDate(transaction.updated_at)}</span>
              </div>
            </div>
          </div>

          {/* Ledger Entry ID */}
          {transaction.ledger_entry_id && (
            <div>
              <p className="text-sm font-medium text-[var(--text-muted)] mb-2">Ledger Entry ID</p>
              <code className="text-sm font-mono text-[var(--text-primary)] bg-[var(--surface-secondary)] px-3 py-2 rounded block break-all">
                {transaction.ledger_entry_id}
              </code>
            </div>
          )}

          {/* Parent Transaction (for reversals/refunds) */}
          {transaction.parent_transaction_id && (
            <div>
              <p className="text-sm font-medium text-[var(--text-muted)] mb-2">Parent Transaction ID</p>
              <code className="text-sm font-mono text-[var(--text-primary)] bg-[var(--surface-secondary)] px-3 py-2 rounded block break-all">
                {transaction.parent_transaction_id}
              </code>
            </div>
          )}

          {/* Failure Reason */}
          {transaction.failure_reason && (
            <div className="bg-[var(--color-error-50)] rounded-lg p-4">
              <p className="text-sm font-medium text-[var(--color-error-800)] mb-2">Failure Reason</p>
              <p className="text-[var(--color-error-900)]">{transaction.failure_reason}</p>
            </div>
          )}

          {/* Metadata */}
          {transaction.metadata && Object.keys(transaction.metadata).length > 0 && (
            <div>
              <p className="text-sm font-medium text-[var(--text-muted)] mb-2">Metadata</p>
              <div className="bg-[var(--surface-secondary)] rounded-lg p-4 space-y-2">
                {Object.entries(transaction.metadata).map(([key, value]) => (
                  <div key={key} className="flex justify-between text-sm">
                    <span className="text-[var(--text-muted)] font-medium">{key}:</span>
                    <span className="text-[var(--text-primary)]">{value}</span>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>

        {/* Footer Actions */}
        <div className="sticky bottom-0 bg-[var(--surface-secondary)] border-t border-[var(--border-subtle)] px-6 py-4 flex justify-end">
          <Button variant="secondary" onClick={onClose}>
            Close
          </Button>
        </div>
      </Card>
    </div>
  );
}
