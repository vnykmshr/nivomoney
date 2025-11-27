import { useState } from 'react';

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
    <div className="card mb-6">
      {/* Filter Header */}
      <div className="flex items-center justify-between mb-4">
        <div className="flex items-center space-x-3">
          <h3 className="text-lg font-semibold text-gray-900">Filter Transactions</h3>
          {hasActiveFilters && (
            <span className="px-2 py-1 bg-primary-100 text-primary-800 text-xs font-semibold rounded-full">
              Active
            </span>
          )}
        </div>
        <div className="flex items-center space-x-2">
          {hasActiveFilters && (
            <button
              onClick={onReset}
              className="text-sm text-primary-600 hover:text-primary-700 font-medium"
            >
              Clear All
            </button>
          )}
          <button
            onClick={() => setIsExpanded(!isExpanded)}
            className="text-gray-500 hover:text-gray-700"
          >
            <svg
              className={`w-5 h-5 transition-transform ${isExpanded ? 'rotate-180' : ''}`}
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
            </svg>
          </button>
        </div>
      </div>

      {/* Search Bar (Always Visible) */}
      <div className="relative">
        <input
          type="text"
          placeholder="Search by description or reference..."
          value={filters.search}
          onChange={e => handleChange('search', e.target.value)}
          className="input-field pl-10"
        />
        <svg
          className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
        </svg>
      </div>

      {/* Expanded Filters */}
      {isExpanded && (
        <div className="mt-4 pt-4 border-t grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {/* Transaction Type */}
          <div>
            <label htmlFor="filter-type" className="block text-sm font-medium text-gray-700 mb-2">
              Transaction Type
            </label>
            <select
              id="filter-type"
              value={filters.type}
              onChange={e => handleChange('type', e.target.value)}
              className="input-field"
            >
              <option value="">All Types</option>
              <option value="deposit">Deposit</option>
              <option value="withdrawal">Withdrawal</option>
              <option value="transfer">Transfer</option>
              <option value="reversal">Reversal</option>
            </select>
          </div>

          {/* Status */}
          <div>
            <label htmlFor="filter-status" className="block text-sm font-medium text-gray-700 mb-2">
              Status
            </label>
            <select
              id="filter-status"
              value={filters.status}
              onChange={e => handleChange('status', e.target.value)}
              className="input-field"
            >
              <option value="">All Statuses</option>
              <option value="pending">Pending</option>
              <option value="processing">Processing</option>
              <option value="completed">Completed</option>
              <option value="failed">Failed</option>
              <option value="reversed">Reversed</option>
            </select>
          </div>

          {/* Date From */}
          <div>
            <label htmlFor="filter-date-from" className="block text-sm font-medium text-gray-700 mb-2">
              From Date
            </label>
            <input
              type="date"
              id="filter-date-from"
              value={filters.dateFrom}
              onChange={e => handleChange('dateFrom', e.target.value)}
              max={filters.dateTo || new Date().toISOString().split('T')[0]}
              className="input-field"
            />
          </div>

          {/* Date To */}
          <div>
            <label htmlFor="filter-date-to" className="block text-sm font-medium text-gray-700 mb-2">
              To Date
            </label>
            <input
              type="date"
              id="filter-date-to"
              value={filters.dateTo}
              onChange={e => handleChange('dateTo', e.target.value)}
              min={filters.dateFrom}
              max={new Date().toISOString().split('T')[0]}
              className="input-field"
            />
          </div>

          {/* Min Amount */}
          <div>
            <label htmlFor="filter-amount-min" className="block text-sm font-medium text-gray-700 mb-2">
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
              className="input-field"
            />
          </div>

          {/* Max Amount */}
          <div>
            <label htmlFor="filter-amount-max" className="block text-sm font-medium text-gray-700 mb-2">
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
              className="input-field"
            />
          </div>
        </div>
      )}

      {/* Quick Filters */}
      {!isExpanded && (
        <div className="mt-4 flex flex-wrap gap-2">
          <button
            onClick={() => handleChange('type', filters.type === 'deposit' ? '' : 'deposit')}
            className={`px-3 py-1 text-sm rounded-full transition-colors ${
              filters.type === 'deposit'
                ? 'bg-primary-600 text-white'
                : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
            }`}
          >
            Deposits
          </button>
          <button
            onClick={() => handleChange('type', filters.type === 'withdrawal' ? '' : 'withdrawal')}
            className={`px-3 py-1 text-sm rounded-full transition-colors ${
              filters.type === 'withdrawal'
                ? 'bg-primary-600 text-white'
                : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
            }`}
          >
            Withdrawals
          </button>
          <button
            onClick={() => handleChange('type', filters.type === 'transfer' ? '' : 'transfer')}
            className={`px-3 py-1 text-sm rounded-full transition-colors ${
              filters.type === 'transfer'
                ? 'bg-primary-600 text-white'
                : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
            }`}
          >
            Transfers
          </button>
          <button
            onClick={() => handleChange('status', filters.status === 'completed' ? '' : 'completed')}
            className={`px-3 py-1 text-sm rounded-full transition-colors ${
              filters.status === 'completed'
                ? 'bg-green-600 text-white'
                : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
            }`}
          >
            Completed
          </button>
          <button
            onClick={() => handleChange('status', filters.status === 'pending' ? '' : 'pending')}
            className={`px-3 py-1 text-sm rounded-full transition-colors ${
              filters.status === 'pending'
                ? 'bg-yellow-600 text-white'
                : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
            }`}
          >
            Pending
          </button>
        </div>
      )}
    </div>
  );
}
