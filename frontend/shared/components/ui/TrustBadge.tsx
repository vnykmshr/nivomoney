import { forwardRef } from 'react';
import type { HTMLAttributes, ReactNode } from 'react';
import { cn } from '../../lib/utils';

export interface TrustBadgeProps extends HTMLAttributes<HTMLDivElement> {
  /** Badge variant */
  variant: 'security' | 'encrypted' | 'verified' | 'instant' | 'custom';
  /** Badge size */
  size?: 'sm' | 'md';
  /** Visual style */
  theme?: 'light' | 'dark';
  /** Custom label (for custom variant) */
  label?: string;
  /** Custom icon (for custom variant) */
  icon?: ReactNode;
  /** Show pulse animation on indicator dot */
  showPulse?: boolean;
}

const ShieldIcon = ({ className }: { className?: string }) => (
  <svg
    className={className}
    fill="none"
    viewBox="0 0 24 24"
    stroke="currentColor"
    strokeWidth={2}
    aria-hidden="true"
  >
    <path
      strokeLinecap="round"
      strokeLinejoin="round"
      d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z"
    />
  </svg>
);

const LockIcon = ({ className }: { className?: string }) => (
  <svg
    className={className}
    fill="none"
    viewBox="0 0 24 24"
    stroke="currentColor"
    strokeWidth={2}
    aria-hidden="true"
  >
    <path
      strokeLinecap="round"
      strokeLinejoin="round"
      d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"
    />
  </svg>
);

const CheckBadgeIcon = ({ className }: { className?: string }) => (
  <svg
    className={className}
    fill="none"
    viewBox="0 0 24 24"
    stroke="currentColor"
    strokeWidth={2}
    aria-hidden="true"
  >
    <path
      strokeLinecap="round"
      strokeLinejoin="round"
      d="M9 12l2 2 4-4M7.835 4.697a3.42 3.42 0 001.946-.806 3.42 3.42 0 014.438 0 3.42 3.42 0 001.946.806 3.42 3.42 0 013.138 3.138 3.42 3.42 0 00.806 1.946 3.42 3.42 0 010 4.438 3.42 3.42 0 00-.806 1.946 3.42 3.42 0 01-3.138 3.138 3.42 3.42 0 00-1.946.806 3.42 3.42 0 01-4.438 0 3.42 3.42 0 00-1.946-.806 3.42 3.42 0 01-3.138-3.138 3.42 3.42 0 00-.806-1.946 3.42 3.42 0 010-4.438 3.42 3.42 0 00.806-1.946 3.42 3.42 0 013.138-3.138z"
    />
  </svg>
);

const BoltIcon = ({ className }: { className?: string }) => (
  <svg
    className={className}
    fill="none"
    viewBox="0 0 24 24"
    stroke="currentColor"
    strokeWidth={2}
    aria-hidden="true"
  >
    <path
      strokeLinecap="round"
      strokeLinejoin="round"
      d="M13 10V3L4 14h7v7l9-11h-7z"
    />
  </svg>
);

const badgeConfig = {
  security: { icon: ShieldIcon, label: 'Bank-Level Security' },
  encrypted: { icon: LockIcon, label: 'End-to-End Encrypted' },
  verified: { icon: CheckBadgeIcon, label: 'Verified & Secure' },
  instant: { icon: BoltIcon, label: 'Instant Transfer' },
  custom: { icon: null, label: '' },
};

export const TrustBadge = forwardRef<HTMLDivElement, TrustBadgeProps>(
  (
    {
      className,
      variant,
      size = 'md',
      theme = 'dark',
      label,
      icon,
      showPulse = true,
      ...props
    },
    ref
  ) => {
    const config = badgeConfig[variant];
    const IconComponent = variant === 'custom' ? () => icon : config.icon;
    const displayLabel = variant === 'custom' ? label : config.label;

    return (
      <div
        ref={ref}
        className={cn(
          'inline-flex items-center gap-2 rounded-full',
          // Size variants
          size === 'sm' && 'px-3 py-1.5 text-xs',
          size === 'md' && 'px-4 py-2 text-sm',
          // Theme variants
          theme === 'dark' && 'bg-white/5 border border-white/10 backdrop-blur-sm',
          theme === 'light' && 'bg-[var(--surface-brand-subtle)] border border-[var(--border-subtle)]',
          className
        )}
        {...props}
      >
        {/* Pulse indicator */}
        <span
          className={cn(
            'rounded-full',
            size === 'sm' && 'w-1.5 h-1.5',
            size === 'md' && 'w-2 h-2',
            showPulse && 'animate-pulse',
            'bg-[var(--color-success-400)]'
          )}
          aria-hidden="true"
        />

        {/* Icon (optional) */}
        {IconComponent && (
          <IconComponent
            className={cn(
              size === 'sm' && 'w-3.5 h-3.5',
              size === 'md' && 'w-4 h-4',
              theme === 'dark' ? 'text-[var(--color-success-400)]' : 'text-[var(--color-success-600)]'
            )}
          />
        )}

        {/* Label */}
        <span
          className={cn(
            theme === 'dark' ? 'text-neutral-300' : 'text-[var(--text-secondary)]'
          )}
        >
          {displayLabel}
        </span>
      </div>
    );
  }
);

TrustBadge.displayName = 'TrustBadge';

// Trust indicator row (multiple badges)
export interface TrustIndicatorRowProps extends HTMLAttributes<HTMLDivElement> {
  items: Array<{
    icon: ReactNode;
    label: string;
  }>;
  theme?: 'light' | 'dark';
}

export const TrustIndicatorRow = forwardRef<HTMLDivElement, TrustIndicatorRowProps>(
  ({ className, items, theme = 'light', ...props }, ref) => {
    return (
      <div
        ref={ref}
        className={cn(
          'flex flex-wrap justify-center gap-6',
          theme === 'dark' ? 'text-neutral-400' : 'text-[var(--text-secondary)]',
          className
        )}
        {...props}
      >
        {items.map((item) => (
          <div key={item.label} className="flex items-center gap-2">
            <span className="text-[var(--color-success-500)]">{item.icon}</span>
            <span className="text-sm">{item.label}</span>
          </div>
        ))}
      </div>
    );
  }
);

TrustIndicatorRow.displayName = 'TrustIndicatorRow';
