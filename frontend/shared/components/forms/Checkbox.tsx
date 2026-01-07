import { forwardRef } from 'react';
import type { InputHTMLAttributes } from 'react';
import { cn } from '../../lib/utils';

export interface CheckboxProps extends Omit<InputHTMLAttributes<HTMLInputElement>, 'type'> {
  label?: string;
  description?: string;
  error?: string;
}

export const Checkbox = forwardRef<HTMLInputElement, CheckboxProps>(
  ({ className, label, description, error, id, ...props }, ref) => {
    const inputId = id || `checkbox-${Math.random().toString(36).slice(2, 9)}`;

    return (
      <div className={cn('relative flex items-start', className)}>
        <div className="flex h-5 items-center">
          <input
            ref={ref}
            type="checkbox"
            id={inputId}
            className={cn(
              'h-4 w-4 rounded-[var(--radius-sm)]',
              'border border-[var(--input-border)]',
              'bg-[var(--input-bg)]',
              'text-[var(--interactive-primary)]',
              'focus:ring-2 focus:ring-[var(--border-focus)] focus:ring-offset-2',
              'focus:ring-offset-[var(--surface-card)]',
              'disabled:opacity-50 disabled:cursor-not-allowed',
              'transition-colors',
              error && 'border-[var(--border-error)]'
            )}
            aria-invalid={!!error}
            aria-describedby={
              error
                ? `${inputId}-error`
                : description
                  ? `${inputId}-description`
                  : undefined
            }
            {...props}
          />
        </div>
        {(label || description || error) && (
          <div className="ml-3 text-sm">
            {label && (
              <label
                htmlFor={inputId}
                className={cn(
                  'font-medium',
                  props.disabled
                    ? 'text-[var(--text-muted)]'
                    : 'text-[var(--text-primary)]'
                )}
              >
                {label}
              </label>
            )}
            {description && (
              <p
                id={`${inputId}-description`}
                className="text-[var(--text-muted)]"
              >
                {description}
              </p>
            )}
            {error && (
              <p
                id={`${inputId}-error`}
                className="text-[var(--text-error)]"
                role="alert"
              >
                {error}
              </p>
            )}
          </div>
        )}
      </div>
    );
  }
);

Checkbox.displayName = 'Checkbox';
