import type { Transaction } from '../types';
import { formatCurrency, formatDate, getStatusColor } from '../lib/utils';

interface TransactionDetailsModalProps {
  transaction: Transaction | null;
  walletId?: string;
  onClose: () => void;
}

export function TransactionDetailsModal({ transaction, walletId, onClose }: TransactionDetailsModalProps) {
  if (!transaction) return null;

  const isIncoming =
    transaction.type === 'deposit' ||
    (transaction.type === 'transfer' && transaction.destination_wallet_id === walletId);

  const getTransactionTypeLabel = (type: string) => {
    const labels: Record<string, string> = {
      deposit: 'Deposit',
      withdrawal: 'Withdrawal',
      transfer: isIncoming ? 'Received Transfer' : 'Sent Transfer',
      reversal: 'Reversal',
    };
    return labels[type] || type;
  };

  const getTransactionIcon = (type: string) => {
    const icons: Record<string, string> = {
      deposit: '↓',
      withdrawal: '↑',
      transfer: isIncoming ? '↓' : '↑',
      reversal: '↺',
    };
    return icons[type] || '•';
  };

  const getAmountDisplay = () => {
    const prefix = isIncoming ? '+' : '-';
    return `${prefix}${formatCurrency(transaction.amount)}`;
  };

  const getAmountColor = () => {
    return isIncoming ? 'text-green-600' : 'text-red-600';
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
      <div className="bg-white rounded-lg max-w-2xl w-full max-h-[90vh] overflow-y-auto">
        {/* Header */}
        <div className="flex items-center justify-between p-6 border-b sticky top-0 bg-white">
          <h2 className="text-2xl font-bold text-gray-900">Transaction Details</h2>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-gray-600 transition-colors"
          >
            <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        {/* Content */}
        <div className="p-6 space-y-6">
          {/* Transaction Type & Amount */}
          <div className="text-center pb-6 border-b">
            <div className="w-16 h-16 rounded-full bg-gradient-to-br from-purple-600 to-primary-600 flex items-center justify-center text-white font-bold text-3xl mx-auto mb-4">
              {getTransactionIcon(transaction.type)}
            </div>
            <h3 className="text-lg font-semibold text-gray-700 mb-2">
              {getTransactionTypeLabel(transaction.type)}
            </h3>
            <div className={`text-4xl font-bold ${getAmountColor()}`}>
              {getAmountDisplay()}
            </div>
            <div className="mt-3">
              <span className={`inline-block px-3 py-1 text-sm font-semibold rounded-full ${getStatusColor(transaction.status)}`}>
                {transaction.status.toUpperCase()}
              </span>
            </div>
          </div>

          {/* Transaction Details Grid */}
          <div className="space-y-4">
            <DetailRow label="Transaction ID" value={transaction.id} mono />
            <DetailRow label="Type" value={transaction.type.charAt(0).toUpperCase() + transaction.type.slice(1)} />
            <DetailRow label="Amount" value={formatCurrency(transaction.amount)} />
            <DetailRow label="Currency" value={transaction.currency} />
            <DetailRow label="Date & Time" value={formatDate(transaction.created_at)} />

            {transaction.description && (
              <DetailRow label="Description" value={transaction.description} />
            )}

            {transaction.reference && (
              <DetailRow label="Reference" value={transaction.reference} mono />
            )}

            {transaction.source_wallet_id && (
              <DetailRow label="Source Wallet ID" value={transaction.source_wallet_id} mono />
            )}

            {transaction.destination_wallet_id && (
              <DetailRow label="Destination Wallet ID" value={transaction.destination_wallet_id} mono />
            )}

            {transaction.parent_transaction_id && (
              <DetailRow label="Parent Transaction ID" value={transaction.parent_transaction_id} mono />
            )}

            <DetailRow label="Last Updated" value={formatDate(transaction.updated_at)} />
          </div>

          {/* Status Information */}
          <div className="pt-6 border-t">
            <h4 className="font-semibold text-gray-900 mb-3">Status Information</h4>
            <div className="bg-gray-50 rounded-lg p-4 space-y-2 text-sm">
              {transaction.status === 'completed' && (
                <p className="text-green-800">
                  ✓ This transaction has been completed successfully.
                </p>
              )}
              {transaction.status === 'pending' && (
                <p className="text-yellow-800">
                  ⏳ This transaction is pending and will be processed soon.
                </p>
              )}
              {transaction.status === 'processing' && (
                <p className="text-blue-800">
                  ⚙️ This transaction is currently being processed.
                </p>
              )}
              {transaction.status === 'failed' && (
                <p className="text-red-800">
                  ✗ This transaction failed. Please contact support if you have questions.
                </p>
              )}
              {transaction.status === 'reversed' && (
                <p className="text-gray-800">
                  ↺ This transaction has been reversed.
                </p>
              )}
            </div>
          </div>

          {/* Actions */}
          <div className="pt-6 border-t flex gap-3">
            <button
              onClick={onClose}
              className="flex-1 btn-primary"
            >
              Close
            </button>
            {/* Future: Add "Download Receipt" button */}
          </div>

          {/* Help Text */}
          <div className="text-center text-sm text-gray-500">
            Need help? Contact support at support@nivomoney.com
          </div>
        </div>
      </div>
    </div>
  );
}

// Helper component for detail rows
interface DetailRowProps {
  label: string;
  value: string;
  mono?: boolean;
}

function DetailRow({ label, value, mono = false }: DetailRowProps) {
  return (
    <div className="flex justify-between items-start py-2 border-b border-gray-100">
      <span className="text-sm font-medium text-gray-600">{label}</span>
      <span className={`text-sm text-gray-900 text-right max-w-xs break-all ${mono ? 'font-mono' : ''}`}>
        {value}
      </span>
    </div>
  );
}
