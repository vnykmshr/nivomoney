import { forwardRef } from 'react';
import type { HTMLAttributes, ReactNode } from 'react';
import { cn } from '../../lib/utils';
import { Button } from './Button';

export interface SuccessStateDetail {
  label: string;
  value: string;
}

export interface SuccessStateAction {
  label: string;
  onClick: () => void;
  variant?: 'primary' | 'secondary';
}

export interface SuccessStateProps extends HTMLAttributes<HTMLDivElement> {
  /** Main success title */
  title: string;
  /** Description/message below title */
  message?: string;
  /** Icon variant */
  icon?: 'check' | 'money' | 'transfer' | 'custom';
  /** Custom icon element (when icon='custom') */
  customIcon?: ReactNode;
  /** Show animated ping effect */
  showAnimation?: boolean;
  /** Details list (label/value pairs) */
  details?: SuccessStateDetail[];
  /** Primary action button */
  primaryAction?: SuccessStateAction;
  /** Secondary action button */
  secondaryAction?: SuccessStateAction;
  /** Icon size */
  size?: 'sm' | 'md' | 'lg';
}

const CheckIcon = ({ className }: { className?: string }) => (
  <svg
    className={className}
    fill="none"
    viewBox="0 0 24 24"
    stroke="currentColor"
    strokeWidth={2.5}
    aria-hidden="true"
  >
    <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" />
  </svg>
);

const MoneyIcon = ({ className }: { className?: string }) => (
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
      d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
    />
  </svg>
);

const TransferIcon = ({ className }: { className?: string }) => (
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
      d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4"
    />
  </svg>
);

export const SuccessState = forwardRef<HTMLDivElement, SuccessStateProps>(
  (
    {
      className,
      title,
      message,
      icon = 'check',
      customIcon,
      showAnimation = true,
      details,
      primaryAction,
      secondaryAction,
      size = 'md',
      ...props
    },
    ref
  ) => {
    const IconComponent = {
      check: CheckIcon,
      money: MoneyIcon,
      transfer: TransferIcon,
      custom: () => customIcon,
    }[icon];

    const iconSizes = {
      sm: 'w-8 h-8',
      md: 'w-12 h-12',
      lg: 'w-14 h-14',
    };

    const containerSizes = {
      sm: 'w-16 h-16',
      md: 'w-24 h-24',
      lg: 'w-28 h-28',
    };

    return (
      <div
        ref={ref}
        className={cn('text-center py-8', className)}
        role="status"
        aria-live="polite"
        {...props}
      >
        {/* Animated Success Icon */}
        <div className="relative mx-auto mb-6" style={{ width: 'fit-content' }}>
          {/* Ping animation ring */}
          {showAnimation && (
            <div
              className={cn(
                'absolute inset-0 rounded-full animate-ping',
                'bg-[var(--color-success-500)]/20'
              )}
              style={{ animationDuration: '1.5s' }}
              aria-hidden="true"
            />
          )}

          {/* Icon container */}
          <div
            className={cn(
              'relative rounded-full flex items-center justify-center',
              'bg-[var(--surface-success)]',
              containerSizes[size]
            )}
          >
            <IconComponent
              className={cn(iconSizes[size], 'text-[var(--color-success-600)]')}
            />
          </div>
        </div>

        {/* Title */}
        <h2 className="text-2xl font-bold text-[var(--text-primary)] mb-2">
          {title}
        </h2>

        {/* Message */}
        {message && (
          <p className="text-lg text-[var(--text-secondary)] mb-6 max-w-md mx-auto">
            {message}
          </p>
        )}

        {/* Details Card */}
        {details && details.length > 0 && (
          <div className="bg-[var(--surface-page)] rounded-xl p-6 text-left mb-8 max-w-sm mx-auto">
            <h3 className="font-semibold text-[var(--text-primary)] mb-4 text-sm uppercase tracking-wide">
              Transaction Details
            </h3>
            <dl className="space-y-3">
              {details.map(({ label, value }) => (
                <div key={label} className="flex justify-between text-sm">
                  <dt className="text-[var(--text-muted)]">{label}</dt>
                  <dd className="font-medium text-[var(--text-primary)]">
                    {value}
                  </dd>
                </div>
              ))}
            </dl>
          </div>
        )}

        {/* Actions */}
        {(primaryAction || secondaryAction) && (
          <div className="flex gap-3 justify-center">
            {secondaryAction && (
              <Button
                variant={secondaryAction.variant || 'secondary'}
                onClick={secondaryAction.onClick}
                className="min-w-[140px]"
              >
                {secondaryAction.label}
              </Button>
            )}
            {primaryAction && (
              <Button
                variant={primaryAction.variant || 'primary'}
                onClick={primaryAction.onClick}
                className="min-w-[140px]"
              >
                {primaryAction.label}
              </Button>
            )}
          </div>
        )}
      </div>
    );
  }
);

SuccessState.displayName = 'SuccessState';
