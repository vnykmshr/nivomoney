import type { Wallet } from '@nivo/shared';
import { formatCurrency } from '../lib/utils';
import { Badge } from '@nivo/shared';

type BadgeVariant = 'success' | 'warning' | 'error' | 'info' | 'neutral';

interface WalletCardProps {
  wallet: Wallet;
  isSelected: boolean;
  onClick: () => void;
}

export function WalletCard({ wallet, isSelected, onClick }: WalletCardProps) {
  const getStatusVariant = (status: string): BadgeVariant => {
    const variants: Record<string, BadgeVariant> = {
      active: 'success',
      inactive: 'neutral',
      frozen: 'warning',
      closed: 'error',
    };
    return variants[status] || 'neutral';
  };

  return (
    <div
      onClick={onClick}
      className={`card cursor-pointer transition-all hover:shadow-md ${
        isSelected ? 'ring-2 ring-[var(--interactive-primary)] bg-[var(--surface-brand-subtle)]' : ''
      }`}
      role="button"
      tabIndex={0}
      aria-pressed={isSelected}
    >
      <div className="flex justify-between items-start mb-3">
        <div className="text-sm font-medium text-[var(--text-secondary)]">
          {wallet.currency} Wallet
        </div>
        <Badge variant={getStatusVariant(wallet.status)}>
          {wallet.status}
        </Badge>
      </div>

      <div className="mb-2">
        <div className="text-sm text-[var(--text-secondary)]">Balance</div>
        <div className="text-2xl font-bold text-[var(--text-primary)]">{formatCurrency(wallet.balance)}</div>
      </div>

      <div>
        <div className="text-sm text-[var(--text-secondary)]">Available Balance</div>
        <div className="text-lg font-semibold text-[var(--text-secondary)]">
          {formatCurrency(wallet.available_balance)}
        </div>
      </div>

      <div className="mt-3 pt-3 border-t border-[var(--border-default)]">
        <div className="text-xs text-[var(--text-muted)]">
          {wallet.currency} • ID: {wallet.id.slice(0, 8)}...
        </div>
      </div>
    </div>
  );
}
