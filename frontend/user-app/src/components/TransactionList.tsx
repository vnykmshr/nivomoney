import { useState } from 'react';
import type { Transaction } from '../types';
import { formatCurrency, formatDate, getStatusColor } from '../lib/utils';
import { TransactionDetailsModal } from './TransactionDetailsModal';

interface TransactionListProps {
  transactions: Transaction[];
  walletId?: string;
}

export function TransactionList({ transactions, walletId }: TransactionListProps) {
  const [selectedTransaction, setSelectedTransaction] = useState<Transaction | null>(null);
  const getTransactionIcon = (type: string) => {
    const icons: Record<string, string> = {
      deposit: '↓',
      withdrawal: '↑',
      transfer: '⇄',
      reversal: '↺',
    };
    return icons[type] || '•';
  };

  const getTransactionAmount = (transaction: Transaction) => {
    const isIncoming =
      transaction.type === 'deposit' ||
      (transaction.type === 'transfer' && transaction.destination_wallet_id === walletId);

    const prefix = isIncoming ? '+' : '-';
    return `${prefix}${formatCurrency(transaction.amount)}`;
  };

  const getAmountColor = (transaction: Transaction) => {
    const isIncoming =
      transaction.type === 'deposit' ||
      (transaction.type === 'transfer' && transaction.destination_wallet_id === walletId);

    return isIncoming ? 'text-green-600' : 'text-red-600';
  };

  if (transactions.length === 0) {
    return (
      <div className="card">
        <div className="text-center py-8 text-gray-500">
          <p className="text-lg">No transactions yet</p>
          <p className="text-sm mt-2">Your transaction history will appear here</p>
        </div>
      </div>
    );
  }

  return (
    <>
      <div className="card">
        <h3 className="text-lg font-semibold mb-4">Recent Transactions</h3>
        <div className="space-y-3">
          {transactions.map(transaction => (
            <div
              key={transaction.id}
              onClick={() => setSelectedTransaction(transaction)}
              className="flex items-center justify-between p-3 bg-gray-50 rounded-lg hover:bg-gray-100 transition-colors cursor-pointer"
            >
            <div className="flex items-center space-x-3">
              <div className="w-10 h-10 rounded-full bg-primary-100 flex items-center justify-center text-primary-700 font-bold text-lg">
                {getTransactionIcon(transaction.type)}
              </div>
              <div>
                <div className="font-medium text-gray-900">
                  {transaction.type.charAt(0).toUpperCase() + transaction.type.slice(1)}
                </div>
                <div className="text-sm text-gray-500">{formatDate(transaction.created_at)}</div>
                {transaction.description && (
                  <div className="text-xs text-gray-400 mt-1">{transaction.description}</div>
                )}
              </div>
            </div>

            <div className="text-right">
              <div className={`font-semibold ${getAmountColor(transaction)}`}>
                {getTransactionAmount(transaction)}
              </div>
              <span
                className={`inline-block px-2 py-1 text-xs font-semibold rounded mt-1 ${getStatusColor(transaction.status)}`}
              >
                {transaction.status}
              </span>
            </div>
          </div>
        ))}
      </div>
    </div>

    {/* Transaction Details Modal */}
    <TransactionDetailsModal
      transaction={selectedTransaction}
      walletId={walletId}
      onClose={() => setSelectedTransaction(null)}
    />
  </>
  );
}
