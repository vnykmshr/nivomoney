import { forwardRef } from 'react';
import type { HTMLAttributes } from 'react';
import { cn } from '../../lib/utils';

export interface AmountDisplayProps extends HTMLAttributes<HTMLDivElement> {
  /** Amount in the smallest unit (paise for INR) */
  amount: number;
  /** Currency symbol */
  currency?: string;
  /** Display size */
  size?: 'sm' | 'md' | 'lg' | 'xl' | '2xl';
  /** Label above the amount */
  label?: string;
  /** Sublabel below the amount */
  sublabel?: string;
  /** Visual variant */
  variant?: 'default' | 'success' | 'highlight' | 'muted';
  /** Whether to show background */
  showBackground?: boolean;
}

/**
 * Formats amount from paise to rupees with Indian number formatting
 */
function formatAmount(amountInPaise: number): string {
  const rupees = amountInPaise / 100;
  return rupees.toLocaleString('en-IN', {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  });
}

export const AmountDisplay = forwardRef<HTMLDivElement, AmountDisplayProps>(
  (
    {
      className,
      amount,
      currency = '₹',
      size = 'lg',
      label,
      sublabel,
      variant = 'default',
      showBackground = false,
      ...props
    },
    ref
  ) => {
    return (
      <div
        ref={ref}
        className={cn(
          'text-center',
          showBackground && [
            'rounded-2xl p-6',
            variant === 'default' && 'bg-[var(--surface-page)]',
            variant === 'success' && 'bg-[var(--surface-success)]',
            variant === 'highlight' && 'bg-gradient-to-br from-[var(--surface-brand-subtle)] to-[var(--surface-page)]',
            variant === 'muted' && 'bg-[var(--surface-muted)]',
          ],
          className
        )}
        {...props}
      >
        {label && (
          <p
            className={cn(
              'mb-1',
              size === 'sm' && 'text-xs',
              size === 'md' && 'text-xs',
              size === 'lg' && 'text-sm',
              size === 'xl' && 'text-sm',
              size === '2xl' && 'text-base',
              variant === 'muted' ? 'text-[var(--text-muted)]' : 'text-[var(--text-secondary)]'
            )}
          >
            {label}
          </p>
        )}

        <p
          className={cn(
            'font-bold tracking-tight tabular-nums',
            // Size variants
            size === 'sm' && 'text-xl',
            size === 'md' && 'text-2xl',
            size === 'lg' && 'text-4xl',
            size === 'xl' && 'text-5xl',
            size === '2xl' && 'text-6xl',
            // Color variants
            variant === 'default' && 'text-[var(--text-primary)]',
            variant === 'success' && 'text-[var(--color-success-600)]',
            variant === 'highlight' && 'text-[var(--interactive-primary)]',
            variant === 'muted' && 'text-[var(--text-muted)]'
          )}
        >
          {currency}
          {formatAmount(amount)}
        </p>

        {sublabel && (
          <p
            className={cn(
              'mt-1',
              size === 'sm' && 'text-xs',
              size === 'md' && 'text-xs',
              size === 'lg' && 'text-sm',
              size === 'xl' && 'text-sm',
              size === '2xl' && 'text-base',
              'text-[var(--text-secondary)]'
            )}
          >
            {sublabel}
          </p>
        )}
      </div>
    );
  }
);

AmountDisplay.displayName = 'AmountDisplay';
