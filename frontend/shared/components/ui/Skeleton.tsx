import type { HTMLAttributes } from 'react';
import { cn } from '../../lib/utils';

export interface SkeletonProps extends HTMLAttributes<HTMLDivElement> {
  variant?: 'text' | 'circular' | 'rectangular';
  width?: string | number;
  height?: string | number;
  lines?: number;
}

export function Skeleton({
  className,
  variant = 'text',
  width,
  height,
  lines = 1,
  style,
  ...props
}: SkeletonProps) {
  const baseClasses = cn(
    'animate-pulse',
    'bg-[var(--interactive-secondary)]',
    variant === 'text' && 'rounded-[var(--radius-md)] h-4',
    variant === 'circular' && 'rounded-full',
    variant === 'rectangular' && 'rounded-[var(--radius-lg)]',
    className
  );

  if (lines > 1 && variant === 'text') {
    return (
      <div className="space-y-2" {...props}>
        {Array.from({ length: lines }).map((_, i) => (
          <div
            key={i}
            className={baseClasses}
            style={{
              width: i === lines - 1 ? '75%' : width || '100%',
              height,
              ...style,
            }}
          />
        ))}
      </div>
    );
  }

  return (
    <div
      className={baseClasses}
      style={{
        width,
        height,
        ...style,
      }}
      {...props}
    />
  );
}

export function SkeletonCard({ className }: { className?: string }) {
  return (
    <div
      className={cn(
        'p-4 rounded-[var(--radius-card)]',
        'bg-[var(--card-bg)]',
        'border border-[var(--card-border)]',
        className
      )}
    >
      <div className="flex items-center gap-3 mb-4">
        <Skeleton variant="circular" width={40} height={40} />
        <div className="flex-1">
          <Skeleton width="60%" className="mb-2" />
          <Skeleton width="40%" />
        </div>
      </div>
      <Skeleton lines={3} />
    </div>
  );
}

export function SkeletonAvatar({
  size = 'md',
  className,
}: {
  size?: 'sm' | 'md' | 'lg';
  className?: string;
}) {
  const sizes = {
    sm: 32,
    md: 40,
    lg: 56,
  };

  return (
    <Skeleton
      variant="circular"
      width={sizes[size]}
      height={sizes[size]}
      className={className}
    />
  );
}
