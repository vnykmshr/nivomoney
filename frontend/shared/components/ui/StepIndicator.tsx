import { forwardRef } from 'react';
import type { HTMLAttributes } from 'react';
import { cn } from '../../lib/utils';

export interface StepIndicatorProps extends HTMLAttributes<HTMLDivElement> {
  /** Array of step labels */
  steps: string[];
  /** Current active step (0-indexed) */
  currentStep: number;
  /** Display variant */
  variant?: 'dots' | 'numbered';
  /** Size variant */
  size?: 'sm' | 'md';
}

const CheckIcon = () => (
  <svg
    className="w-4 h-4"
    fill="none"
    viewBox="0 0 24 24"
    stroke="currentColor"
    strokeWidth={3}
    aria-hidden="true"
  >
    <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" />
  </svg>
);

export const StepIndicator = forwardRef<HTMLDivElement, StepIndicatorProps>(
  (
    {
      className,
      steps,
      currentStep,
      variant = 'numbered',
      size = 'md',
      ...props
    },
    ref
  ) => {
    return (
      <div
        ref={ref}
        className={cn('flex items-center justify-center', className)}
        role="navigation"
        aria-label="Progress"
        {...props}
      >
        <ol className="flex items-center gap-2">
          {steps.map((step, index) => {
            const isCompleted = index < currentStep;
            const isCurrent = index === currentStep;
            const isPending = index > currentStep;

            return (
              <li key={step} className="flex items-center">
                {/* Step Circle */}
                <div
                  className={cn(
                    'flex items-center justify-center rounded-full font-medium transition-all',
                    // Size
                    size === 'sm' && 'w-6 h-6 text-xs',
                    size === 'md' && 'w-8 h-8 text-sm',
                    // State colors
                    isCompleted && 'bg-[var(--color-success-500)] text-white',
                    isCurrent && 'bg-[var(--interactive-primary)] text-white shadow-lg shadow-[var(--color-primary-500)]/20',
                    isPending && 'bg-[var(--surface-muted)] text-[var(--text-muted)]'
                  )}
                  aria-current={isCurrent ? 'step' : undefined}
                >
                  {isCompleted ? (
                    <CheckIcon />
                  ) : variant === 'numbered' ? (
                    index + 1
                  ) : (
                    <span
                      className={cn(
                        'rounded-full',
                        size === 'sm' && 'w-2 h-2',
                        size === 'md' && 'w-2.5 h-2.5',
                        isCurrent ? 'bg-white' : 'bg-current'
                      )}
                    />
                  )}
                </div>

                {/* Connector Line (not after last step) */}
                {index < steps.length - 1 && (
                  <div
                    className={cn(
                      'mx-2 transition-colors',
                      size === 'sm' && 'w-6 h-0.5',
                      size === 'md' && 'w-8 h-0.5',
                      isCompleted
                        ? 'bg-[var(--color-success-500)]'
                        : 'bg-[var(--border-subtle)]'
                    )}
                    aria-hidden="true"
                  />
                )}
              </li>
            );
          })}
        </ol>
      </div>
    );
  }
);

StepIndicator.displayName = 'StepIndicator';

// Compact version with labels
export interface StepIndicatorWithLabelsProps extends StepIndicatorProps {
  /** Show step labels below indicators */
  showLabels?: boolean;
}

export const StepIndicatorWithLabels = forwardRef<
  HTMLDivElement,
  StepIndicatorWithLabelsProps
>(
  (
    {
      className,
      steps,
      currentStep,
      variant = 'numbered',
      size = 'md',
      showLabels = true,
      ...props
    },
    ref
  ) => {
    return (
      <div ref={ref} className={cn('w-full', className)} {...props}>
        <nav aria-label="Progress">
          <ol className="flex items-center justify-between">
            {steps.map((step, index) => {
              const isCompleted = index < currentStep;
              const isCurrent = index === currentStep;
              const isPending = index > currentStep;

              return (
                <li
                  key={step}
                  className={cn('flex flex-col items-center', index < steps.length - 1 && 'flex-1')}
                >
                  <div className="flex items-center w-full">
                    {/* Step Circle */}
                    <div
                      className={cn(
                        'flex items-center justify-center rounded-full font-medium transition-all shrink-0',
                        size === 'sm' && 'w-6 h-6 text-xs',
                        size === 'md' && 'w-8 h-8 text-sm',
                        isCompleted && 'bg-[var(--color-success-500)] text-white',
                        isCurrent && 'bg-[var(--interactive-primary)] text-white shadow-lg shadow-[var(--color-primary-500)]/20',
                        isPending && 'bg-[var(--surface-muted)] text-[var(--text-muted)]'
                      )}
                      aria-current={isCurrent ? 'step' : undefined}
                    >
                      {isCompleted ? (
                        <CheckIcon />
                      ) : variant === 'numbered' ? (
                        index + 1
                      ) : null}
                    </div>

                    {/* Connector Line */}
                    {index < steps.length - 1 && (
                      <div
                        className={cn(
                          'flex-1 mx-2 h-0.5 transition-colors',
                          isCompleted
                            ? 'bg-[var(--color-success-500)]'
                            : 'bg-[var(--border-subtle)]'
                        )}
                        aria-hidden="true"
                      />
                    )}
                  </div>

                  {/* Label */}
                  {showLabels && (
                    <span
                      className={cn(
                        'mt-2 text-xs text-center',
                        isCurrent
                          ? 'text-[var(--text-primary)] font-medium'
                          : 'text-[var(--text-muted)]'
                      )}
                    >
                      {step}
                    </span>
                  )}
                </li>
              );
            })}
          </ol>
        </nav>
      </div>
    );
  }
);

StepIndicatorWithLabels.displayName = 'StepIndicatorWithLabels';
