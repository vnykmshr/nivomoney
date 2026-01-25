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
  Skeleton,
} from '@nivo/shared';
import {
  cn,
  getTransactionStatusVariant,
  getTransactionTypeVariant,
} from '@nivo/shared';

export function Transactions() {
  const [searchQuery, setSearchQuery] = useState('');
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [isSearching, setIsSearching] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [selectedTransactionId, setSelectedTransactionId] = useState<string | null>(null);

  // Filters
  const [statusFilter, setStatusFilter] = useState<string>('');
  const [typeFilter, setTypeFilter] = useState<string>('');

  // Status filter options for chips
  const statusOptions = [
    { value: '', label: 'All' },
    { value: 'completed', label: 'Completed' },
    { value: 'pending', label: 'Pending' },
    { value: 'failed', label: 'Failed' },
  ];

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

  // Escape CSV field: quote if contains comma, quote, or newline
  const escapeCSVField = (field: string | undefined | null): string => {
    if (!field) return '';
    const str = String(field);
    // If field contains special chars, quote it and escape internal quotes
    if (str.includes(',') || str.includes('"') || str.includes('\n') || str.includes('\r')) {
      return `"${str.replace(/"/g, '""').replace(/\r?\n/g, ' ')}"`;
    }
    return str;
  };

  const handleExportCSV = () => {
    if (transactions.length === 0) {
      setError('No transactions to export');
      return;
    }

    const headers = ['ID', 'Type', 'Status', 'Amount', 'Currency', 'Description', 'Reference', 'Created At'];
    const rows = transactions.map(tx => [
      tx.id,
      tx.type,
      tx.status,
      (tx.amount / 100).toFixed(2),
      tx.currency,
      escapeCSVField(tx.description),
      escapeCSVField(tx.reference),
      new Date(tx.created_at).toISOString(),
    ]);

    const csvContent = [
      headers.join(','),
      ...rows.map(row => row.join(','))
    ].join('\n');

    const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
    const link = document.createElement('a');
    link.href = URL.createObjectURL(blob);
    link.download = `transactions_${new Date().toISOString().split('T')[0]}.csv`;
    link.click();
    URL.revokeObjectURL(link.href);
  };

  const clearFilters = () => {
    setSearchQuery('');
    setStatusFilter('');
    setTypeFilter('');
  };

  const hasActiveFilters = searchQuery || statusFilter || typeFilter;

  return (
    <AdminLayout title="Transaction Search">
      <div className="space-y-6">
        {/* Error Alert */}
        {error && (
          <Alert variant="error" onDismiss={() => setError(null)}>
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

            {/* Status Filter Chips */}
            <div>
              <label className="block text-sm font-medium text-[var(--text-secondary)] mb-2">
                Status
              </label>
              <div className="flex flex-wrap gap-2">
                {statusOptions.map(option => (
                  <button
                    key={option.value}
                    onClick={() => setStatusFilter(option.value)}
                    className={cn(
                      'px-3 py-1.5 rounded-full text-sm font-medium transition-colors',
                      statusFilter === option.value
                        ? 'bg-[var(--interactive-primary)] text-white'
                        : 'bg-[var(--surface-secondary)] text-[var(--text-secondary)] hover:bg-[var(--surface-tertiary)]'
                    )}
                  >
                    {option.label}
                  </button>
                ))}
              </div>
            </div>

            {/* Type Filter */}
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

            {/* Action Buttons */}
            <div className="flex gap-3">
              <Button
                onClick={handleSearch}
                loading={isSearching}
                className="flex-1"
              >
                Search Transactions
              </Button>
              {hasActiveFilters && (
                <Button variant="secondary" onClick={clearFilters}>
                  Clear
                </Button>
              )}
            </div>
          </div>
        </Card>

        {/* Loading State */}
        {isSearching && (
          <div className="space-y-4">
            {[1, 2, 3].map(i => (
              <Card key={i}>
                <div className="flex items-center gap-3 mb-2">
                  <Skeleton className="h-6 w-20" />
                  <Skeleton className="h-6 w-20" />
                </div>
                <Skeleton className="h-8 w-32 mb-2" />
                <Skeleton className="h-4 w-full mb-1" />
                <Skeleton className="h-3 w-48 mt-2" />
              </Card>
            ))}
          </div>
        )}

        {/* Results */}
        {!isSearching && transactions.length > 0 && (
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <h3 className="text-lg font-semibold text-[var(--text-primary)]">
                Found {transactions.length} transaction{transactions.length !== 1 ? 's' : ''}
              </h3>
              <Button variant="secondary" size="sm" onClick={handleExportCSV}>
                <svg className="w-4 h-4 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
                </svg>
                Export CSV
              </Button>
            </div>

            {transactions.map((tx) => (
              <Card
                key={tx.id}
                className="cursor-pointer hover:shadow-md transition-shadow"
                onClick={() => setSelectedTransactionId(tx.id)}
              >
                <div className="flex justify-between items-start">
                  <div className="flex-1">
                    <div className="flex items-center gap-3 mb-2">
                      <Badge variant={getTransactionTypeVariant(tx.type)}>
                        {tx.type}
                      </Badge>
                      <Badge variant={getTransactionStatusVariant(tx.status)}>
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
        )}

        {/* Empty State */}
        {!isSearching && transactions.length === 0 && (
          <Card className="text-center py-12">
            <div className="w-16 h-16 mx-auto mb-4 rounded-full bg-[var(--surface-secondary)] flex items-center justify-center">
              <svg className="w-8 h-8 text-[var(--text-muted)]" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
              </svg>
            </div>
            <p className="text-[var(--text-muted)]">Enter search criteria to find transactions</p>
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
