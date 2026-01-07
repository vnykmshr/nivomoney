import { cn } from '../../lib/utils';

export interface SpinnerProps {
  size?: 'sm' | 'md' | 'lg';
  className?: string;
}

export function Spinner({ size = 'md', className }: SpinnerProps) {
  const sizes = {
    sm: 'w-4 h-4',
    md: 'w-6 h-6',
    lg: 'w-8 h-8',
  };

  return (
    <svg
      className={cn('animate-spin text-[var(--interactive-primary)]', sizes[size], className)}
      fill="none"
      viewBox="0 0 24 24"
      role="status"
      aria-label="Loading"
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
  );
}

export interface LoadingOverlayProps {
  visible: boolean;
  message?: string;
  className?: string;
}

export function LoadingOverlay({ visible, message, className }: LoadingOverlayProps) {
  if (!visible) return null;

  return (
    <div
      className={cn(
        'absolute inset-0 z-10',
        'flex flex-col items-center justify-center gap-3',
        'bg-[var(--surface-card)]/80 backdrop-blur-sm',
        'rounded-[inherit]',
        className
      )}
    >
      <Spinner size="lg" />
      {message && (
        <p className="text-sm text-[var(--text-secondary)]">{message}</p>
      )}
    </div>
  );
}
