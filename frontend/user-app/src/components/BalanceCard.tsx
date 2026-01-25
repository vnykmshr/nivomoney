import { cn } from '@nivo/shared';
import type { Wallet } from '@nivo/shared';

export interface BalanceCardProps {
  wallet: Wallet;
  totalBalance: number;
  availableBalance: number;
  className?: string;
}

export function BalanceCard({
  wallet,
  totalBalance,
  availableBalance,
  className,
}: BalanceCardProps) {
  const formatCurrency = (amount: number): string => {
    return new Intl.NumberFormat('en-IN', {
      style: 'currency',
      currency: wallet.currency || 'INR',
      minimumFractionDigits: 2,
    }).format(amount / 100);
  };

  const heldBalance = totalBalance - availableBalance;

  return (
    <div
      className={cn(
        'relative overflow-hidden rounded-2xl p-6',
        'bg-gradient-to-br from-[var(--color-primary-500)] via-[var(--color-primary-600)] to-[var(--color-primary-700)]',
        'text-white shadow-lg',
        className
      )}
    >
      {/* Background decoration */}
      <div className="absolute top-0 right-0 w-40 h-40 -mr-10 -mt-10 rounded-full bg-white/10" aria-hidden="true" />
      <div className="absolute bottom-0 left-0 w-32 h-32 -ml-8 -mb-8 rounded-full bg-white/5" aria-hidden="true" />

      {/* Content */}
      <div className="relative z-10">
        {/* Wallet type label */}
        <div className="flex items-center gap-2 mb-4">
          <div className="w-10 h-10 rounded-full bg-white/20 flex items-center justify-center">
            <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 10h18M7 15h1m4 0h1m-7 4h12a3 3 0 003-3V8a3 3 0 00-3-3H6a3 3 0 00-3 3v8a3 3 0 003 3z" />
            </svg>
          </div>
          <div>
            <p className="text-sm text-white/80">
              {wallet.type === 'default' ? 'Primary' : wallet.type} Wallet
            </p>
            <p className="text-xs text-white/60 font-mono">
              {wallet.id.slice(0, 8)}...
            </p>
          </div>
        </div>

        {/* Main balance */}
        <div className="mb-6">
          <p className="text-sm text-white/80 mb-1">Available Balance</p>
          <p className="text-4xl font-bold tracking-tight">
            {formatCurrency(availableBalance)}
          </p>
        </div>

        {/* Secondary info */}
        <div className="flex items-center gap-6 pt-4 border-t border-white/20">
          <div>
            <p className="text-xs text-white/60 mb-0.5">Total Balance</p>
            <p className="text-lg font-semibold">{formatCurrency(totalBalance)}</p>
          </div>
          {heldBalance > 0 && (
            <div>
              <p className="text-xs text-white/60 mb-0.5">On Hold</p>
              <p className="text-lg font-semibold text-[var(--color-warning-300)]">
                {formatCurrency(heldBalance)}
              </p>
            </div>
          )}
          <div className="ml-auto">
            <span
              className={cn(
                'inline-flex items-center gap-1 px-2 py-1 rounded-full text-xs font-medium',
                wallet.status === 'active'
                  ? 'bg-[var(--color-success-500)]/20 text-[var(--color-success-200)]'
                  : 'bg-white/20 text-white/80'
              )}
            >
              <span
                className={cn(
                  'w-1.5 h-1.5 rounded-full',
                  wallet.status === 'active' ? 'bg-[var(--color-success-400)]' : 'bg-white/60'
                )}
              />
              {wallet.status === 'active' ? 'Active' : wallet.status}
            </span>
          </div>
        </div>
      </div>
    </div>
  );
}
