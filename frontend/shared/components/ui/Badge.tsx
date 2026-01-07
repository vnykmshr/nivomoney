import type { HTMLAttributes } from 'react';
import { cn } from '../../lib/utils';

export interface BadgeProps extends HTMLAttributes<HTMLSpanElement> {
  variant?: 'success' | 'warning' | 'error' | 'info' | 'neutral';
  size?: 'sm' | 'md';
}

export function Badge({
  className,
  variant = 'neutral',
  size = 'md',
  children,
  ...props
}: BadgeProps) {
  return (
    <span
      className={cn(
        'inline-flex items-center font-medium',
        'rounded-[var(--radius-badge)]',

        // Size variants
        size === 'sm' && 'px-2 py-0.5 text-xs',
        size === 'md' && 'px-2.5 py-1 text-xs',

        // Color variants
        variant === 'success' && [
          'bg-[var(--surface-success)]',
          'text-[var(--text-success)]',
        ],
        variant === 'warning' && [
          'bg-[var(--surface-warning)]',
          'text-[var(--text-warning)]',
        ],
        variant === 'error' && [
          'bg-[var(--surface-error)]',
          'text-[var(--text-error)]',
        ],
        variant === 'info' && [
          'bg-[var(--surface-brand-subtle)]',
          'text-[var(--text-link)]',
        ],
        variant === 'neutral' && [
          'bg-[var(--interactive-secondary)]',
          'text-[var(--text-secondary)]',
        ],

        className
      )}
      {...props}
    >
      {children}
    </span>
  );
}
