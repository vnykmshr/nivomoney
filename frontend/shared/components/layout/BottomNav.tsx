import type { ReactNode } from 'react';
import { cn } from '../../lib/utils';

export interface BottomNavProps {
  children: ReactNode;
  className?: string;
}

export function BottomNav({ children, className }: BottomNavProps) {
  return (
    <nav
      className={cn(
        'fixed bottom-0 left-0 right-0 z-40',
        'md:hidden',
        'bg-[var(--surface-card)]',
        'border-t border-[var(--border-subtle)]',
        'safe-area-inset-bottom',
        className
      )}
    >
      <div className="flex items-center justify-around h-16">{children}</div>
    </nav>
  );
}

export interface BottomNavItemProps {
  href: string;
  icon: ReactNode;
  label: string;
  active?: boolean;
  badge?: string | number;
  className?: string;
}

export function BottomNavItem({
  href,
  icon,
  label,
  active = false,
  badge,
  className,
}: BottomNavItemProps) {
  return (
    <a
      href={href}
      aria-current={active ? 'page' : undefined}
      className={cn(
        'flex flex-col items-center justify-center',
        'min-w-[64px] py-2 px-3',
        'text-xs font-medium',
        'transition-colors',
        active
          ? 'text-[var(--interactive-primary)]'
          : 'text-[var(--text-muted)] hover:text-[var(--text-secondary)]',
        className
      )}
    >
      <span className="relative mb-1">
        <span className="w-6 h-6 flex items-center justify-center">{icon}</span>
        {badge !== undefined && (
          <span
            className={cn(
              'absolute -top-1 -right-1',
              'min-w-[16px] h-4 px-1',
              'flex items-center justify-center',
              'text-[10px] font-bold',
              'rounded-full',
              'bg-[var(--interactive-danger)] text-[var(--text-inverse)]'
            )}
          >
            {badge}
          </span>
        )}
      </span>
      <span>{label}</span>
    </a>
  );
}
