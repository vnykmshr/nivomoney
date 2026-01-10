import { forwardRef } from 'react';
import type { HTMLAttributes, ReactNode } from 'react';
import { cn } from '../../lib/utils';

export interface PageHeroProps extends HTMLAttributes<HTMLElement> {
  variant?: 'dark' | 'gradient' | 'light';
  size?: 'sm' | 'md' | 'lg';
  showGlow?: boolean;
  showGrid?: boolean;
  children?: ReactNode;
}

const gridPatternSvg = `url("data:image/svg+xml,%3Csvg width='60' height='60' viewBox='0 0 60 60' xmlns='http://www.w3.org/2000/svg'%3E%3Cg fill='none' fill-rule='evenodd'%3E%3Cg fill='%23ffffff' fill-opacity='1'%3E%3Cpath d='M36 34v-4h-2v4h-4v2h4v4h2v-4h4v-2h-4zm0-30V0h-2v4h-4v2h4v4h2V6h4V4h-4zM6 34v-4H4v4H0v2h4v4h2v-4h4v-2H6zM6 4V0H4v4H0v2h4v4h2V6h4V4H6z'/%3E%3C/g%3E%3C/g%3E%3C/svg%3E")`;

export const PageHero = forwardRef<HTMLElement, PageHeroProps>(
  (
    {
      className,
      variant = 'dark',
      size = 'md',
      showGlow = true,
      showGrid = true,
      children,
      ...props
    },
    ref
  ) => {
    return (
      <section
        ref={ref}
        className={cn(
          'relative overflow-hidden',

          // Variant backgrounds
          variant === 'dark' && 'bg-gradient-to-br from-neutral-900 via-neutral-950 to-neutral-900 text-white',
          variant === 'gradient' && 'bg-gradient-to-br from-[var(--color-primary-500)] via-[var(--color-primary-600)] to-[var(--color-primary-700)] text-white',
          variant === 'light' && 'bg-[var(--surface-page)] text-[var(--text-primary)]',

          // Size variants
          size === 'sm' && 'py-8 md:py-12',
          size === 'md' && 'py-12 md:py-20',
          size === 'lg' && 'py-16 md:py-28',

          className
        )}
        {...props}
      >
        {/* Teal Glow Effects (dark/gradient variants only) */}
        {showGlow && variant !== 'light' && (
          <>
            <div
              className="absolute top-0 right-0 w-[500px] h-[500px] rounded-full blur-3xl -translate-y-1/2 translate-x-1/4 pointer-events-none"
              style={{ backgroundColor: 'rgba(0, 172, 176, 0.1)' }}
              aria-hidden="true"
            />
            <div
              className="absolute bottom-0 left-0 w-[400px] h-[400px] rounded-full blur-3xl translate-y-1/3 -translate-x-1/4 pointer-events-none"
              style={{ backgroundColor: 'rgba(0, 140, 142, 0.1)' }}
              aria-hidden="true"
            />
          </>
        )}

        {/* Grid Pattern Overlay (dark variant only) */}
        {showGrid && variant === 'dark' && (
          <div
            className="absolute inset-0 opacity-[0.02] pointer-events-none"
            style={{ backgroundImage: gridPatternSvg }}
            aria-hidden="true"
          />
        )}

        {/* Content */}
        <div className="relative z-10">{children}</div>
      </section>
    );
  }
);

PageHero.displayName = 'PageHero';

// Sub-component for wave separator
export interface WaveSeparatorProps extends HTMLAttributes<HTMLDivElement> {
  fillColor?: string;
}

export const WaveSeparator = forwardRef<HTMLDivElement, WaveSeparatorProps>(
  ({ className, fillColor = 'var(--surface-card)', ...props }, ref) => {
    return (
      <div ref={ref} className={cn('relative h-16', className)} {...props}>
        <svg
          className="absolute bottom-0 w-full h-16"
          preserveAspectRatio="none"
          viewBox="0 0 1440 74"
          fill="none"
          aria-hidden="true"
        >
          <path
            d="M0 74V0C240 49.3333 480 74 720 74C960 74 1200 49.3333 1440 0V74H0Z"
            fill={fillColor}
          />
        </svg>
      </div>
    );
  }
);

WaveSeparator.displayName = 'WaveSeparator';
