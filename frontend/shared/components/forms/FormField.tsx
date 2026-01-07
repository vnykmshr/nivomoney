import type { ReactNode } from 'react';
import { cn } from '../../lib/utils';

export interface FormFieldProps {
  label?: string;
  htmlFor?: string;
  error?: string;
  hint?: string;
  required?: boolean;
  children: ReactNode;
  className?: string;
}

export function FormField({
  label,
  htmlFor,
  error,
  hint,
  required,
  children,
  className,
}: FormFieldProps) {
  return (
    <div className={cn('space-y-1.5', className)}>
      {label && (
        <label
          htmlFor={htmlFor}
          className="block text-sm font-medium text-[var(--text-primary)]"
        >
          {label}
          {required && (
            <span className="text-[var(--text-error)] ml-0.5" aria-hidden="true">
              *
            </span>
          )}
        </label>
      )}
      {children}
      {error && (
        <p className="text-sm text-[var(--text-error)]" role="alert">
          {error}
        </p>
      )}
      {hint && !error && (
        <p className="text-sm text-[var(--text-muted)]">{hint}</p>
      )}
    </div>
  );
}
