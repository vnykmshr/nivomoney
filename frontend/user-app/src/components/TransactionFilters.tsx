import { useState } from 'react';
import { Card, Input, Badge } from '@nivo/shared';
import { cn } from '@nivo/shared';

export interface TransactionFilterValues {
  search: string;
  type: string;
  status: string;
  dateFrom: string;
  dateTo: string;
  amountMin: string;
  amountMax: string;
}

interface TransactionFiltersProps {
  filters: TransactionFilterValues;
  onFilterChange: (filters: TransactionFilterValues) => void;
  onReset: () => void;
}

export function TransactionFilters({ filters, onFilterChange, onReset }: TransactionFiltersProps) {
  const [isExpanded, setIsExpanded] = useState(false);

  const handleChange = (field: keyof TransactionFilterValues, value: string) => {
    onFilterChange({
      ...filters,
      [field]: value,
    });
  };

  const hasActiveFilters = Object.values(filters).some(v => v !== '');

  return (
    <Card className="mb-6">
      {/* Filter Header */}
      <div className="flex items-center justify-between mb-4">
        <div className="flex items-center space-x-3">
          <h3 className="text-lg font-semibold text-[var(--text-primary)]">Filter Transactions</h3>
          {hasActiveFilters && (
            <Badge variant="info">Active</Badge>
          )}
        </div>
        <div className="flex items-center space-x-2">
          {hasActiveFilters && (
            <button
              onClick={onReset}
              className="text-sm text-[var(--interactive-primary)] hover:text-[var(--interactive-primary-hover)] font-medium"
            >
              Clear All
            </button>
          )}
          <button
            onClick={() => setIsExpanded(!isExpanded)}
            className="text-[var(--text-muted)] hover:text-[var(--text-primary)]"
            aria-label={isExpanded ? 'Collapse filters' : 'Expand filters'}
            aria-expanded={isExpanded}
          >
            <svg
              className={cn('w-5 h-5 transition-transform', isExpanded && 'rotate-180')}
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
              aria-hidden="true"
            >
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
            </svg>
          </button>
        </div>
      </div>

      {/* Search Bar (Always Visible) */}
      <div className="relative">
        <label htmlFor="transaction-search" className="sr-only">Search transactions</label>
        <Input
          id="transaction-search"
          type="text"
          placeholder="Search by description or reference..."
          value={filters.search}
          onChange={e => handleChange('search', e.target.value)}
          className="pl-10"
        />
        <svg
          className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-[var(--text-muted)]"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
          aria-hidden="true"
        >
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
        </svg>
      </div>

      {/* Expanded Filters */}
      {isExpanded && (
        <div className="mt-4 pt-4 border-t border-[var(--border-subtle)] grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {/* Transaction Type */}
          <div>
            <label htmlFor="filter-type" className="block text-sm font-medium text-[var(--text-secondary)] mb-2">
              Transaction Type
            </label>
            <select
              id="filter-type"
              value={filters.type}
              onChange={e => handleChange('type', e.target.value)}
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
          </div>

          {/* Status */}
          <div>
            <label htmlFor="filter-status" className="block text-sm font-medium text-[var(--text-secondary)] mb-2">
              Status
            </label>
            <select
              id="filter-status"
              value={filters.status}
              onChange={e => handleChange('status', e.target.value)}
              className="w-full px-4 py-3 rounded-lg border bg-[var(--surface-input)] border-[var(--border-default)] text-[var(--text-primary)] focus:outline-none focus:ring-2 focus:ring-[var(--interactive-primary)] focus:border-transparent"
            >
              <option value="">All Statuses</option>
              <option value="pending">Pending</option>
              <option value="processing">Processing</option>
              <option value="completed">Completed</option>
              <option value="failed">Failed</option>
              <option value="reversed">Reversed</option>
              <option value="cancelled">Cancelled</option>
            </select>
          </div>

          {/* Date From */}
          <div>
            <label htmlFor="filter-date-from" className="block text-sm font-medium text-[var(--text-secondary)] mb-2">
              From Date
            </label>
            <input
              type="date"
              id="filter-date-from"
              value={filters.dateFrom}
              onChange={e => handleChange('dateFrom', e.target.value)}
              max={filters.dateTo || new Date().toISOString().split('T')[0]}
              className="w-full px-4 py-3 rounded-lg border bg-[var(--surface-input)] border-[var(--border-default)] text-[var(--text-primary)] focus:outline-none focus:ring-2 focus:ring-[var(--interactive-primary)] focus:border-transparent"
            />
          </div>

          {/* Date To */}
          <div>
            <label htmlFor="filter-date-to" className="block text-sm font-medium text-[var(--text-secondary)] mb-2">
              To Date
            </label>
            <input
              type="date"
              id="filter-date-to"
              value={filters.dateTo}
              onChange={e => handleChange('dateTo', e.target.value)}
              min={filters.dateFrom}
              max={new Date().toISOString().split('T')[0]}
              className="w-full px-4 py-3 rounded-lg border bg-[var(--surface-input)] border-[var(--border-default)] text-[var(--text-primary)] focus:outline-none focus:ring-2 focus:ring-[var(--interactive-primary)] focus:border-transparent"
            />
          </div>

          {/* Min Amount */}
          <div>
            <label htmlFor="filter-amount-min" className="block text-sm font-medium text-[var(--text-secondary)] mb-2">
              Min Amount (₹)
            </label>
            <input
              type="number"
              id="filter-amount-min"
              value={filters.amountMin}
              onChange={e => handleChange('amountMin', e.target.value)}
              min="0"
              step="0.01"
              placeholder="0.00"
              className="w-full px-4 py-3 rounded-lg border bg-[var(--surface-input)] border-[var(--border-default)] text-[var(--text-primary)] focus:outline-none focus:ring-2 focus:ring-[var(--interactive-primary)] focus:border-transparent"
            />
          </div>

          {/* Max Amount */}
          <div>
            <label htmlFor="filter-amount-max" className="block text-sm font-medium text-[var(--text-secondary)] mb-2">
              Max Amount (₹)
            </label>
            <input
              type="number"
              id="filter-amount-max"
              value={filters.amountMax}
              onChange={e => handleChange('amountMax', e.target.value)}
              min={filters.amountMin || "0"}
              step="0.01"
              placeholder="0.00"
              className="w-full px-4 py-3 rounded-lg border bg-[var(--surface-input)] border-[var(--border-default)] text-[var(--text-primary)] focus:outline-none focus:ring-2 focus:ring-[var(--interactive-primary)] focus:border-transparent"
            />
          </div>
        </div>
      )}

      {/* Quick Filters */}
      {!isExpanded && (
        <div className="mt-4 flex flex-wrap gap-2">
          <button
            onClick={() => handleChange('type', filters.type === 'deposit' ? '' : 'deposit')}
            className={cn(
              'px-3 py-1 text-sm rounded-full transition-colors',
              filters.type === 'deposit'
                ? 'bg-[var(--interactive-primary)] text-white'
                : 'bg-[var(--surface-secondary)] text-[var(--text-secondary)] hover:bg-[var(--interactive-secondary)]'
            )}
          >
            Deposits
          </button>
          <button
            onClick={() => handleChange('type', filters.type === 'withdrawal' ? '' : 'withdrawal')}
            className={cn(
              'px-3 py-1 text-sm rounded-full transition-colors',
              filters.type === 'withdrawal'
                ? 'bg-[var(--interactive-primary)] text-white'
                : 'bg-[var(--surface-secondary)] text-[var(--text-secondary)] hover:bg-[var(--interactive-secondary)]'
            )}
          >
            Withdrawals
          </button>
          <button
            onClick={() => handleChange('type', filters.type === 'transfer' ? '' : 'transfer')}
            className={cn(
              'px-3 py-1 text-sm rounded-full transition-colors',
              filters.type === 'transfer'
                ? 'bg-[var(--interactive-primary)] text-white'
                : 'bg-[var(--surface-secondary)] text-[var(--text-secondary)] hover:bg-[var(--interactive-secondary)]'
            )}
          >
            Transfers
          </button>
          <button
            onClick={() => handleChange('status', filters.status === 'completed' ? '' : 'completed')}
            className={cn(
              'px-3 py-1 text-sm rounded-full transition-colors',
              filters.status === 'completed'
                ? 'bg-[var(--color-success-600)] text-white'
                : 'bg-[var(--surface-secondary)] text-[var(--text-secondary)] hover:bg-[var(--interactive-secondary)]'
            )}
          >
            Completed
          </button>
          <button
            onClick={() => handleChange('status', filters.status === 'pending' ? '' : 'pending')}
            className={cn(
              'px-3 py-1 text-sm rounded-full transition-colors',
              filters.status === 'pending'
                ? 'bg-[var(--color-warning-600)] text-white'
                : 'bg-[var(--surface-secondary)] text-[var(--text-secondary)] hover:bg-[var(--interactive-secondary)]'
            )}
          >
            Pending
          </button>
        </div>
      )}
    </Card>
  );
}
