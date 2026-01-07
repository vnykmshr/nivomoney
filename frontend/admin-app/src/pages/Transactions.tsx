/**
 * Transactions Page - Admin Transaction Monitoring
 * Global transaction search and monitoring for admins
 */

import { useState } from 'react';
import { AdminLayout } from '../components';
import { adminApi } from '../lib/adminApi';
import { TransactionDetailModal } from '../components/TransactionDetailModal';
import type { Transaction } from '@nivo/shared';
import {
  Card,
  CardTitle,
  Button,
  Input,
  Alert,
  Badge,
  FormField,
} from '../../../shared/components';

type BadgeVariant = 'success' | 'warning' | 'error' | 'info' | 'neutral';

export function Transactions() {
  const [searchQuery, setSearchQuery] = useState('');
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [isSearching, setIsSearching] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [selectedTransactionId, setSelectedTransactionId] = useState<string | null>(null);

  // Filters
  const [statusFilter, setStatusFilter] = useState<string>('');
  const [typeFilter, setTypeFilter] = useState<string>('');

  const handleSearch = async () => {
    if (!searchQuery.trim() && !statusFilter && !typeFilter) {
      setError('Please enter a search term or select a filter');
      return;
    }

    setIsSearching(true);
    setError(null);

    try {
      const results = await adminApi.searchTransactions({
        search: searchQuery.trim() || undefined,
        status: statusFilter || undefined,
        type: typeFilter || undefined,
      });
      setTransactions(results);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Search failed');
      setTransactions([]);
    } finally {
      setIsSearching(false);
    }
  };

  const getStatusVariant = (status: string): BadgeVariant => {
    switch (status) {
      case 'completed': return 'success';
      case 'pending': return 'warning';
      case 'failed': return 'error';
      case 'processing': return 'info';
      case 'reversed': return 'info';
      case 'cancelled': return 'neutral';
      default: return 'neutral';
    }
  };

  const getTypeVariant = (type: string): BadgeVariant => {
    switch (type) {
      case 'deposit': return 'success';
      case 'withdrawal': return 'warning';
      case 'transfer': return 'info';
      case 'reversal': return 'info';
      case 'fee': return 'error';
      case 'refund': return 'success';
      default: return 'neutral';
    }
  };

  return (
    <AdminLayout title="Transaction Search">
      <div className="space-y-6">
        {/* Error Alert */}
        {error && (
          <Alert variant="error" dismissible onDismiss={() => setError(null)}>
            {error}
          </Alert>
        )}

        {/* Search Card */}
        <Card>
          <CardTitle className="mb-4">Search Transactions</CardTitle>

          <div className="space-y-4">
            {/* Search Input */}
            <FormField
              label="Search by transaction ID, description, or reference"
              htmlFor="search-transactions"
            >
              <Input
                id="search-transactions"
                type="text"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                onKeyDown={(e) => e.key === 'Enter' && handleSearch()}
                placeholder="Enter search term..."
              />
            </FormField>

            {/* Filters */}
            <div className="grid grid-cols-2 gap-4">
              <FormField label="Status" htmlFor="status-filter">
                <select
                  id="status-filter"
                  value={statusFilter}
                  onChange={(e) => setStatusFilter(e.target.value)}
                  className="w-full px-4 py-3 rounded-lg border bg-[var(--surface-input)] border-[var(--border-default)] text-[var(--text-primary)] focus:outline-none focus:ring-2 focus:ring-[var(--interactive-primary)] focus:border-transparent"
                >
                  <option value="">All Statuses</option>
                  <option value="completed">Completed</option>
                  <option value="pending">Pending</option>
                  <option value="processing">Processing</option>
                  <option value="failed">Failed</option>
                  <option value="reversed">Reversed</option>
                  <option value="cancelled">Cancelled</option>
                </select>
              </FormField>

              <FormField label="Type" htmlFor="type-filter">
                <select
                  id="type-filter"
                  value={typeFilter}
                  onChange={(e) => setTypeFilter(e.target.value)}
                  className="w-full px-4 py-3 rounded-lg border bg-[var(--surface-input)] border-[var(--border-default)] text-[var(--text-primary)] focus:outline-none focus:ring-2 focus:ring-[var(--interactive-primary)] focus:border-transparent"
                >
                  <option value="">All Types</option>
                  <option value="deposit">Deposit</option>
                  <option value="withdrawal">Withdrawal</option>
                  <option value="transfer">Transfer</option>
                  <option value="reversal">Reversal</option>
                  <option value="fee">Fee</option>
                  <option value="refund">Refund</option>
                </select>
              </FormField>
            </div>

            {/* Search Button */}
            <Button
              onClick={handleSearch}
              loading={isSearching}
              className="w-full"
            >
              Search Transactions
            </Button>
          </div>
        </Card>

        {/* Results */}
        {transactions.length > 0 ? (
          <div className="space-y-4">
            <h3 className="text-lg font-semibold text-[var(--text-primary)]">
              Found {transactions.length} transaction{transactions.length !== 1 ? 's' : ''}
            </h3>

            {transactions.map((tx) => (
              <Card
                key={tx.id}
                className="cursor-pointer hover:shadow-md transition-shadow"
                onClick={() => setSelectedTransactionId(tx.id)}
              >
                <div className="flex justify-between items-start">
                  <div className="flex-1">
                    <div className="flex items-center gap-3 mb-2">
                      <Badge variant={getTypeVariant(tx.type)}>
                        {tx.type}
                      </Badge>
                      <Badge variant={getStatusVariant(tx.status)}>
                        {tx.status}
                      </Badge>
                    </div>

                    <p className="text-2xl font-bold text-[var(--text-primary)] mb-2">
                      {tx.currency} {(tx.amount / 100).toFixed(2)}
                    </p>

                    <p className="text-[var(--text-secondary)] mb-1">{tx.description}</p>

                    {tx.reference && (
                      <p className="text-sm text-[var(--text-secondary)]">Reference: {tx.reference}</p>
                    )}

                    <p className="text-xs text-[var(--text-muted)] mt-2 font-mono">
                      Transaction ID: {tx.id}
                    </p>
                    <p className="text-xs text-[var(--text-muted)]">
                      Created: {new Date(tx.created_at).toLocaleString()}
                    </p>
                  </div>
                </div>
              </Card>
            ))}
          </div>
        ) : (
          <Card className="text-center py-12">
            {isSearching ? (
              <p className="text-[var(--text-muted)]">Searching...</p>
            ) : (
              <>
                <div className="w-16 h-16 mx-auto mb-4 rounded-full bg-[var(--surface-secondary)] flex items-center justify-center">
                  <svg className="w-8 h-8 text-[var(--text-muted)]" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                  </svg>
                </div>
                <p className="text-[var(--text-muted)]">Enter search criteria to find transactions</p>
              </>
            )}
          </Card>
        )}
      </div>

      {/* Transaction Detail Modal */}
      {selectedTransactionId && (
        <TransactionDetailModal
          transactionId={selectedTransactionId}
          onClose={() => setSelectedTransactionId(null)}
        />
      )}
    </AdminLayout>
  );
}
