/**
 * Transactions Page - Admin Transaction Monitoring
 * Global transaction search and monitoring for admins
 */

import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { adminApi } from '../lib/adminApi';
import { TransactionDetailModal } from '../components/TransactionDetailModal';
import type { Transaction } from '@nivo/shared';

export function Transactions() {
  const navigate = useNavigate();
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

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <nav className="bg-white shadow-sm">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between h-16 items-center">
            <div className="flex items-center space-x-4">
              <button
                onClick={() => navigate('/')}
                className="text-gray-600 hover:text-gray-900"
              >
                ← Back
              </button>
              <h1 className="text-xl font-bold text-primary-600">Transaction Search</h1>
            </div>
          </div>
        </div>
      </nav>

      {/* Main Content */}
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Error Alert */}
        {error && (
          <div className="mb-6 p-4 bg-red-100 text-red-800 rounded-lg flex justify-between items-center">
            <span>{error}</span>
            <button onClick={() => setError(null)} className="text-red-600 hover:text-red-800">
              ✕
            </button>
          </div>
        )}

        {/* Search Card */}
        <div className="card mb-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Search Transactions</h2>

          <div className="space-y-4">
            {/* Search Input */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Search by transaction ID, description, or reference
              </label>
              <input
                type="text"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                onKeyPress={(e) => e.key === 'Enter' && handleSearch()}
                placeholder="Enter search term..."
                className="input-field w-full"
              />
            </div>

            {/* Filters */}
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">Status</label>
                <select
                  value={statusFilter}
                  onChange={(e) => setStatusFilter(e.target.value)}
                  className="input-field w-full"
                >
                  <option value="">All Statuses</option>
                  <option value="completed">Completed</option>
                  <option value="pending">Pending</option>
                  <option value="processing">Processing</option>
                  <option value="failed">Failed</option>
                  <option value="reversed">Reversed</option>
                  <option value="cancelled">Cancelled</option>
                </select>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">Type</label>
                <select
                  value={typeFilter}
                  onChange={(e) => setTypeFilter(e.target.value)}
                  className="input-field w-full"
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
            </div>

            {/* Search Button */}
            <button
              onClick={handleSearch}
              disabled={isSearching}
              className="btn-primary w-full"
            >
              {isSearching ? 'Searching...' : 'Search Transactions'}
            </button>
          </div>
        </div>

        {/* Results */}
        {transactions.length > 0 ? (
          <div className="space-y-4">
            <h3 className="text-lg font-semibold text-gray-900">
              Found {transactions.length} transaction{transactions.length !== 1 ? 's' : ''}
            </h3>

            {transactions.map((tx) => (
              <div
                key={tx.id}
                className="card hover:shadow-md transition-shadow cursor-pointer"
                onClick={() => setSelectedTransactionId(tx.id)}
              >
                <div className="flex justify-between items-start">
                  <div className="flex-1">
                    <div className="flex items-center space-x-3 mb-2">
                      <span className={`px-3 py-1 rounded-full text-sm ${getTypeColor(tx.type)}`}>
                        {tx.type}
                      </span>
                      <span className={`px-3 py-1 rounded-full text-sm ${getStatusColor(tx.status)}`}>
                        {tx.status}
                      </span>
                    </div>

                    <p className="text-2xl font-bold text-gray-900 mb-2">
                      {tx.currency} {(tx.amount / 100).toFixed(2)}
                    </p>

                    <p className="text-gray-700 mb-1">{tx.description}</p>

                    {tx.reference && (
                      <p className="text-sm text-gray-600">Reference: {tx.reference}</p>
                    )}

                    <p className="text-xs text-gray-500 mt-2">
                      Transaction ID: {tx.id}
                    </p>
                    <p className="text-xs text-gray-500">
                      Created: {new Date(tx.created_at).toLocaleString()}
                    </p>
                  </div>
                </div>
              </div>
            ))}
          </div>
        ) : (
          <div className="card text-center py-12 text-gray-500">
            {isSearching ? (
              <p>Searching...</p>
            ) : (
              <p>Enter search criteria to find transactions</p>
            )}
          </div>
        )}
      </div>

      {/* Transaction Detail Modal */}
      {selectedTransactionId && (
        <TransactionDetailModal
          transactionId={selectedTransactionId}
          onClose={() => setSelectedTransactionId(null)}
        />
      )}
    </div>
  );
}
