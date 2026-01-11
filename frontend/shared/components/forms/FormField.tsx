import { cloneElement, isValidElement, useId } from 'react';
import type { ReactNode, ReactElement } from 'react';
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
  const generatedId = useId();
  const errorId = error ? `${generatedId}-error` : undefined;
  const hintId = hint && !error ? `${generatedId}-hint` : undefined;

  // Combine IDs for aria-describedby (error takes precedence, hint shown when no error)
  const describedByIds = [errorId, hintId].filter(Boolean).join(' ') || undefined;

  // Clone child input to add accessibility props
  const enhancedChildren = isValidElement(children)
    ? cloneElement(children as ReactElement<Record<string, unknown>>, {
        'aria-required': required || undefined,
        'aria-invalid': error ? 'true' : undefined,
        'aria-describedby': describedByIds,
        error: !!error,
        errorId,
      })
    : children;

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
      {enhancedChildren}
      {error && (
        <p id={errorId} className="text-sm text-[var(--text-error)]" role="alert">
          {error}
        </p>
      )}
      {hint && !error && (
        <p id={hintId} className="text-sm text-[var(--text-muted)]">{hint}</p>
      )}
    </div>
  );
}
