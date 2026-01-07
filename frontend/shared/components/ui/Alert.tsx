import type { HTMLAttributes, ReactNode } from 'react';
import { cn } from '../../lib/utils';

export interface AlertProps extends HTMLAttributes<HTMLDivElement> {
  variant?: 'success' | 'warning' | 'error' | 'info';
  title?: string;
  icon?: ReactNode;
  onDismiss?: () => void;
}

const defaultIcons: Record<string, ReactNode> = {
  success: (
    <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
    </svg>
  ),
  warning: (
    <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
    </svg>
  ),
  error: (
    <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" />
    </svg>
  ),
  info: (
    <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
    </svg>
  ),
};

export function Alert({
  className,
  variant = 'info',
  title,
  icon,
  onDismiss,
  children,
  ...props
}: AlertProps) {
  const displayIcon = icon ?? defaultIcons[variant];

  return (
    <div
      role="alert"
      className={cn(
        'relative flex gap-3 p-4',
        'rounded-[var(--radius-card)]',
        'border',

        // Color variants
        variant === 'success' && [
          'bg-[var(--surface-success)]',
          'border-[var(--color-success-200)]',
          'text-[var(--text-success)]',
        ],
        variant === 'warning' && [
          'bg-[var(--surface-warning)]',
          'border-[var(--color-warning-200)]',
          'text-[var(--text-warning)]',
        ],
        variant === 'error' && [
          'bg-[var(--surface-error)]',
          'border-[var(--color-error-200)]',
          'text-[var(--text-error)]',
        ],
        variant === 'info' && [
          'bg-[var(--surface-brand-subtle)]',
          'border-[var(--color-primary-200)]',
          'text-[var(--text-link)]',
        ],

        className
      )}
      {...props}
    >
      {displayIcon && <div className="flex-shrink-0">{displayIcon}</div>}
      <div className="flex-1 min-w-0">
        {title && (
          <h4 className="font-medium mb-1">{title}</h4>
        )}
        <div className="text-sm opacity-90">{children}</div>
      </div>
      {onDismiss && (
        <button
          onClick={onDismiss}
          className="flex-shrink-0 opacity-70 hover:opacity-100 transition-opacity"
          aria-label="Dismiss"
        >
          <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      )}
    </div>
  );
}
