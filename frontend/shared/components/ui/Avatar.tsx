import { useState } from 'react';
import type { HTMLAttributes } from 'react';
import { cn } from '../../lib/utils';

export interface AvatarProps extends HTMLAttributes<HTMLDivElement> {
  src?: string;
  alt?: string;
  name?: string;
  size?: 'sm' | 'md' | 'lg' | 'xl';
}

export function Avatar({
  src,
  alt,
  name,
  size = 'md',
  className,
  ...props
}: AvatarProps) {
  const [imageError, setImageError] = useState(false);

  const sizes = {
    sm: 'w-8 h-8 text-xs',
    md: 'w-10 h-10 text-sm',
    lg: 'w-14 h-14 text-base',
    xl: 'w-20 h-20 text-xl',
  };

  const getInitials = (name: string): string => {
    return name
      .split(' ')
      .map((part) => part[0])
      .join('')
      .toUpperCase()
      .slice(0, 2);
  };

  const initials = name ? getInitials(name) : '?';
  const showImage = src && !imageError;

  return (
    <div
      className={cn(
        'relative inline-flex items-center justify-center',
        'rounded-full overflow-hidden',
        'bg-[var(--surface-brand)] text-[var(--text-inverse)]',
        'font-medium',
        sizes[size],
        className
      )}
      {...props}
    >
      {showImage ? (
        <img
          src={src}
          alt={alt || name || 'Avatar'}
          className="w-full h-full object-cover"
          onError={() => setImageError(true)}
        />
      ) : (
        <span aria-hidden="true">{initials}</span>
      )}
    </div>
  );
}

export interface AvatarGroupProps {
  children: React.ReactNode;
  max?: number;
  size?: 'sm' | 'md' | 'lg';
  className?: string;
}

export function AvatarGroup({
  children,
  max = 4,
  size = 'md',
  className,
}: AvatarGroupProps) {
  const childArray = Array.isArray(children) ? children : [children];
  const visibleChildren = childArray.slice(0, max);
  const remainingCount = childArray.length - max;

  const overlapSizes = {
    sm: '-ml-2',
    md: '-ml-3',
    lg: '-ml-4',
  };

  return (
    <div className={cn('flex items-center', className)}>
      {visibleChildren.map((child, index) => (
        <div
          key={index}
          className={cn(
            'ring-2 ring-[var(--surface-card)] rounded-full',
            index > 0 && overlapSizes[size]
          )}
        >
          {child}
        </div>
      ))}
      {remainingCount > 0 && (
        <div
          className={cn(
            'flex items-center justify-center rounded-full',
            'bg-[var(--interactive-secondary)] text-[var(--text-secondary)]',
            'ring-2 ring-[var(--surface-card)]',
            'font-medium text-xs',
            overlapSizes[size],
            size === 'sm' && 'w-8 h-8',
            size === 'md' && 'w-10 h-10',
            size === 'lg' && 'w-14 h-14'
          )}
        >
          +{remainingCount}
        </div>
      )}
    </div>
  );
}
