import { forwardRef } from 'react';
import type { ButtonHTMLAttributes } from 'react';
import { cn } from '../../lib/utils';

export interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary' | 'outline' | 'danger' | 'ghost';
  size?: 'sm' | 'md' | 'lg';
  loading?: boolean;
}

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  (
    {
      className,
      variant = 'primary',
      size = 'md',
      loading,
      children,
      disabled,
      ...props
    },
    ref
  ) => {
    return (
      <button
        ref={ref}
        className={cn(
          // Base styles
          'inline-flex items-center justify-center font-medium',
          'rounded-[var(--radius-button)]',
          'transition-[color,background-color,border-color]',
          'duration-150 ease-in-out',
          'focus-visible:outline-none focus-visible:[box-shadow:var(--focus-ring)]',
          'disabled:opacity-50 disabled:cursor-not-allowed',

          // Size variants
          size === 'sm' && 'h-8 px-3 text-sm gap-1.5',
          size === 'md' && 'h-10 px-4 text-sm gap-2',
          size === 'lg' && 'h-12 px-6 text-base gap-2',

          // Color variants - ALL use semantic tokens
          variant === 'primary' && [
            'bg-[var(--button-primary-bg)]',
            'text-[var(--button-primary-text)]',
            'hover:bg-[var(--button-primary-bg-hover)]',
            'shadow-[var(--shadow-button)]',
          ],
          variant === 'secondary' && [
            'bg-[var(--button-secondary-bg)]',
            'text-[var(--button-secondary-text)]',
            'hover:bg-[var(--button-secondary-bg-hover)]',
          ],
          variant === 'outline' && [
            'border-2 border-[var(--interactive-primary)]',
            'text-[var(--interactive-primary)]',
            'bg-transparent',
            'hover:bg-[var(--surface-brand-subtle)]',
          ],
          variant === 'danger' && [
            'bg-[var(--interactive-danger)]',
            'text-[var(--text-inverse)]',
            'hover:bg-[var(--interactive-danger-hover)]',
          ],
          variant === 'ghost' && [
            'text-[var(--text-primary)]',
            'bg-transparent',
            'hover:bg-[var(--interactive-secondary)]',
          ],

          className
        )}
        disabled={disabled || loading}
        aria-busy={loading || undefined}
        {...props}
      >
        {loading && (
          <svg
            className="animate-spin h-4 w-4"
            fill="none"
            viewBox="0 0 24 24"
            aria-hidden="true"
          >
            <circle
              className="opacity-25"
              cx="12"
              cy="12"
              r="10"
              stroke="currentColor"
              strokeWidth="4"
            />
            <path
              className="opacity-75"
              fill="currentColor"
              d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"
            />
          </svg>
        )}
        {children}
      </button>
    );
  }
);

Button.displayName = 'Button';
