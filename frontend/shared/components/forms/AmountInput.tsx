import type { ChangeEvent } from 'react';
import { cn } from '../../lib/utils';

export interface AmountInputProps {
  value: number | '';
  onChange: (value: number | '') => void;
  currency?: string;
  max?: number;
  min?: number;
  label?: string;
  error?: string;
  hint?: string;
  disabled?: boolean;
  placeholder?: string;
  className?: string;
}

export function AmountInput({
  value,
  onChange,
  currency = '₹',
  max,
  min = 0,
  label,
  error,
  hint,
  disabled = false,
  placeholder = '0',
  className,
}: AmountInputProps) {
  const formatAmount = (num: number): string => {
    return new Intl.NumberFormat('en-IN').format(num);
  };

  const handleChange = (e: ChangeEvent<HTMLInputElement>) => {
    // Remove all non-numeric characters except decimal point
    const raw = e.target.value.replace(/[^0-9.]/g, '');

    // Handle empty input
    if (raw === '' || raw === '.') {
      onChange('');
      return;
    }

    // Parse and validate
    const num = parseFloat(raw);
    if (isNaN(num)) return;

    // Check bounds
    if (min !== undefined && num < min) return;
    if (max !== undefined && num > max) return;

    onChange(num);
  };

  return (
    <div className={className}>
      {label && (
        <label className="block text-sm font-medium text-[var(--text-secondary)] mb-1.5">
          {label}
        </label>
      )}
      <div className="relative">
        <span
          className="absolute left-4 top-1/2 -translate-y-1/2 text-2xl text-[var(--text-muted)] select-none"
          aria-hidden="true"
        >
          {currency}
        </span>
        <input
          type="text"
          inputMode="decimal"
          value={value === '' ? '' : formatAmount(value)}
          onChange={handleChange}
          disabled={disabled}
          placeholder={placeholder}
          aria-label={label || 'Amount'}
          aria-invalid={!!error}
          className={cn(
            'w-full pl-12 pr-4 py-4 text-3xl font-semibold',
            'rounded-[var(--radius-input)]',
            'border-2 border-[var(--input-border)]',
            'bg-[var(--input-bg)] text-[var(--text-primary)]',
            'placeholder:text-[var(--text-muted)]',
            'focus:border-[var(--input-border-focus)]',
            'focus:outline-none focus:[box-shadow:var(--focus-ring)]',
            'disabled:opacity-50 disabled:cursor-not-allowed',
            'transition-[border-color,box-shadow] duration-150',
            'tabular-nums',
            error && 'border-[var(--border-error)] focus:[box-shadow:var(--focus-ring-error)]'
          )}
        />
      </div>
      {error && (
        <p className="mt-1.5 text-sm text-[var(--text-error)]" role="alert">
          {error}
        </p>
      )}
      {hint && !error && (
        <p className="mt-1.5 text-sm text-[var(--text-muted)]">{hint}</p>
      )}
    </div>
  );
}
