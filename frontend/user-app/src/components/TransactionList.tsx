import { useState } from 'react';
import type { Transaction } from '@nivo/shared';
import { formatCurrency, formatDate } from '../lib/utils';
import { TransactionDetailsModal } from './TransactionDetailsModal';
import { Badge } from '../../../shared/components';

type BadgeVariant = 'success' | 'warning' | 'error' | 'info' | 'neutral';

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
      fee: '₹',
      refund: '↩',
    };
    return icons[type] || '•';
  };

  const isIncomingTransaction = (transaction: Transaction) => {
    return (
      transaction.type === 'deposit' ||
      transaction.type === 'refund' ||
      (transaction.type === 'transfer' && transaction.destination_wallet_id === walletId)
    );
  };

  const getTransactionAmount = (transaction: Transaction) => {
    const prefix = isIncomingTransaction(transaction) ? '+' : '-';
    return `${prefix}${formatCurrency(transaction.amount)}`;
  };

  const getAmountColor = (transaction: Transaction) => {
    return isIncomingTransaction(transaction)
      ? 'text-[var(--text-success)]'
      : 'text-[var(--text-error)]';
  };

  const getStatusVariant = (status: string): BadgeVariant => {
    const variants: Record<string, BadgeVariant> = {
      completed: 'success',
      pending: 'warning',
      processing: 'info',
      failed: 'error',
      reversed: 'neutral',
    };
    return variants[status] || 'neutral';
  };

  if (transactions.length === 0) {
    return (
      <div className="card">
        <div className="text-center py-8 text-[var(--text-muted)]">
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
              className="flex items-center justify-between p-3 bg-[var(--surface-page)] rounded-lg hover:bg-[var(--interactive-secondary)] transition-colors cursor-pointer"
            >
              <div className="flex items-center space-x-3">
                <div className="w-10 h-10 rounded-full bg-[var(--surface-brand-subtle)] flex items-center justify-center text-[var(--interactive-primary)] font-bold text-lg">
                  {getTransactionIcon(transaction.type)}
                </div>
                <div>
                  <div className="font-medium text-[var(--text-primary)]">
                    {transaction.type.charAt(0).toUpperCase() + transaction.type.slice(1)}
                  </div>
                  <div className="text-sm text-[var(--text-muted)]">{formatDate(transaction.created_at)}</div>
                  {transaction.description && (
                    <div className="text-xs text-[var(--text-muted)] mt-1">{transaction.description}</div>
                  )}
                </div>
              </div>

              <div className="text-right">
                <div className={`font-semibold ${getAmountColor(transaction)}`}>
                  {getTransactionAmount(transaction)}
                </div>
                <Badge variant={getStatusVariant(transaction.status)} className="mt-1">
                  {transaction.status}
                </Badge>
              </div>
            </div>
          ))}
        </div>
      </div>

      <TransactionDetailsModal
        transaction={selectedTransaction}
        walletId={walletId}
        onClose={() => setSelectedTransaction(null)}
      />
    </>
  );
}
