import type { Transaction } from '@nivo/shared';
import { formatCurrency, formatDate } from '../lib/utils';
import { Modal, Badge, Button } from '../../../shared/components';

type BadgeVariant = 'success' | 'warning' | 'error' | 'info' | 'neutral';

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
    return isIncoming ? 'text-[var(--text-success)]' : 'text-[var(--text-error)]';
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

  const getStatusMessage = (status: string) => {
    const messages: Record<string, { icon: string; text: string; color: string }> = {
      completed: { icon: '✓', text: 'This transaction has been completed successfully.', color: 'text-[var(--text-success)]' },
      pending: { icon: '⏳', text: 'This transaction is pending and will be processed soon.', color: 'text-[var(--text-warning)]' },
      processing: { icon: '⚙️', text: 'This transaction is currently being processed.', color: 'text-[var(--text-link)]' },
      failed: { icon: '✗', text: 'This transaction failed. Please contact support if you have questions.', color: 'text-[var(--text-error)]' },
      reversed: { icon: '↺', text: 'This transaction has been reversed.', color: 'text-[var(--text-secondary)]' },
    };
    return messages[status];
  };

  const statusMessage = getStatusMessage(transaction.status);

  return (
    <Modal isOpen={true} onClose={onClose} title="Transaction Details" size="lg">
      <div className="space-y-6">
        {/* Transaction Type & Amount */}
        <div className="text-center pb-6 border-b border-[var(--border-subtle)]">
          <div className="w-16 h-16 rounded-full bg-gradient-to-br from-[var(--color-primary-400)] to-[var(--color-primary-600)] flex items-center justify-center text-[var(--text-inverse)] font-bold text-3xl mx-auto mb-4">
            {getTransactionIcon(transaction.type)}
          </div>
          <h3 className="text-lg font-semibold text-[var(--text-secondary)] mb-2">
            {getTransactionTypeLabel(transaction.type)}
          </h3>
          <div className={`text-4xl font-bold ${getAmountColor()}`}>
            {getAmountDisplay()}
          </div>
          <div className="mt-3">
            <Badge variant={getStatusVariant(transaction.status)}>
              {transaction.status.toUpperCase()}
            </Badge>
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
        {statusMessage && (
          <div className="pt-6 border-t border-[var(--border-subtle)]">
            <h4 className="font-semibold text-[var(--text-primary)] mb-3">Status Information</h4>
            <div className="bg-[var(--surface-page)] rounded-lg p-4 text-sm">
              <p className={statusMessage.color}>
                {statusMessage.icon} {statusMessage.text}
              </p>
            </div>
          </div>
        )}

        {/* Actions */}
        <div className="pt-6 border-t border-[var(--border-subtle)] flex gap-3">
          <Button onClick={onClose} className="flex-1">
            Close
          </Button>
        </div>

        {/* Help Text */}
        <div className="text-center text-sm text-[var(--text-muted)]">
          Need help? Contact support at support@nivomoney.com
        </div>
      </div>
    </Modal>
  );
}

interface DetailRowProps {
  label: string;
  value: string;
  mono?: boolean;
}

function DetailRow({ label, value, mono = false }: DetailRowProps) {
  return (
    <div className="flex justify-between items-start py-2 border-b border-[var(--border-subtle)]">
      <span className="text-sm font-medium text-[var(--text-secondary)]">{label}</span>
      <span className={`text-sm text-[var(--text-primary)] text-right max-w-xs break-all ${mono ? 'font-mono' : ''}`}>
        {value}
      </span>
    </div>
  );
}
