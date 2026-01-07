import { useRef } from 'react';
import type { KeyboardEvent, ClipboardEvent } from 'react';
import { cn } from '../../lib/utils';

export interface OTPInputProps {
  length?: number;
  value: string;
  onChange: (value: string) => void;
  error?: string;
  autoFocus?: boolean;
  disabled?: boolean;
  className?: string;
}

export function OTPInput({
  length = 6,
  value,
  onChange,
  error,
  autoFocus = true,
  disabled = false,
  className,
}: OTPInputProps) {
  const inputRefs = useRef<(HTMLInputElement | null)[]>([]);

  const handleChange = (index: number, digit: string) => {
    if (!/^\d*$/.test(digit)) return;

    const newValue = value.split('');
    newValue[index] = digit;
    const result = newValue.join('').slice(0, length);
    onChange(result);

    // Auto-advance to next input
    if (digit && index < length - 1) {
      inputRefs.current[index + 1]?.focus();
    }
  };

  const handleKeyDown = (index: number, e: KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Backspace') {
      if (!value[index] && index > 0) {
        // Move to previous input on backspace if current is empty
        inputRefs.current[index - 1]?.focus();
      }
    } else if (e.key === 'ArrowLeft' && index > 0) {
      e.preventDefault();
      inputRefs.current[index - 1]?.focus();
    } else if (e.key === 'ArrowRight' && index < length - 1) {
      e.preventDefault();
      inputRefs.current[index + 1]?.focus();
    }
  };

  const handlePaste = (e: ClipboardEvent<HTMLInputElement>) => {
    e.preventDefault();
    const pasted = e.clipboardData.getData('text').replace(/\D/g, '').slice(0, length);
    if (pasted) {
      onChange(pasted);
      // Focus the input after the last pasted digit
      const focusIndex = Math.min(pasted.length, length - 1);
      inputRefs.current[focusIndex]?.focus();
    }
  };

  const handleFocus = (e: React.FocusEvent<HTMLInputElement>) => {
    e.target.select();
  };

  return (
    <div className={className}>
      <div className="flex gap-2 justify-center" role="group" aria-label="One-time password input">
        {Array.from({ length }).map((_, index) => (
          <input
            key={index}
            ref={(el) => {
              inputRefs.current[index] = el;
            }}
            type="text"
            inputMode="numeric"
            autoComplete="one-time-code"
            maxLength={1}
            value={value[index] || ''}
            onChange={(e) => handleChange(index, e.target.value)}
            onKeyDown={(e) => handleKeyDown(index, e)}
            onPaste={handlePaste}
            onFocus={handleFocus}
            autoFocus={autoFocus && index === 0}
            disabled={disabled}
            aria-label={`Digit ${index + 1} of ${length}`}
            className={cn(
              'w-12 h-14 text-center text-2xl font-semibold',
              'rounded-[var(--radius-input)]',
              'border-2 border-[var(--input-border)]',
              'bg-[var(--input-bg)] text-[var(--text-primary)]',
              'focus:border-[var(--input-border-focus)]',
              'focus:outline-none focus:[box-shadow:var(--focus-ring)]',
              'disabled:opacity-50 disabled:cursor-not-allowed',
              'transition-[border-color,box-shadow] duration-150',
              error && 'border-[var(--border-error)] focus:[box-shadow:var(--focus-ring-error)]'
            )}
          />
        ))}
      </div>
      {error && (
        <p className="mt-2 text-sm text-center text-[var(--text-error)]" role="alert">
          {error}
        </p>
      )}
    </div>
  );
}
