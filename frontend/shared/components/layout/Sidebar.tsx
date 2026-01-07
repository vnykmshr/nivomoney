import type { ReactNode } from 'react';
import { cn } from '../../lib/utils';

export interface SidebarProps {
  children: ReactNode;
  collapsed?: boolean;
  className?: string;
}

export function Sidebar({ children, collapsed = false, className }: SidebarProps) {
  return (
    <aside
      className={cn(
        'hidden md:flex flex-col',
        'bg-[var(--surface-card)]',
        'border-r border-[var(--border-subtle)]',
        'transition-[width] duration-200',
        collapsed ? 'w-16' : 'w-64',
        className
      )}
    >
      <nav className="flex-1 p-3 space-y-1">{children}</nav>
    </aside>
  );
}

export interface SidebarItemProps {
  href: string;
  icon: ReactNode;
  label: string;
  active?: boolean;
  collapsed?: boolean;
  badge?: string | number;
  className?: string;
}

export function SidebarItem({
  href,
  icon,
  label,
  active = false,
  collapsed = false,
  badge,
  className,
}: SidebarItemProps) {
  return (
    <a
      href={href}
      aria-current={active ? 'page' : undefined}
      className={cn(
        'flex items-center gap-3 px-3 py-2.5',
        'rounded-[var(--radius-lg)]',
        'text-sm font-medium',
        'transition-colors',
        active
          ? 'bg-[var(--surface-brand-subtle)] text-[var(--interactive-primary)]'
          : 'text-[var(--text-secondary)] hover:text-[var(--text-primary)] hover:bg-[var(--interactive-secondary)]',
        collapsed && 'justify-center px-2',
        className
      )}
      title={collapsed ? label : undefined}
    >
      <span className="w-5 h-5 flex-shrink-0">{icon}</span>
      {!collapsed && (
        <>
          <span className="flex-1">{label}</span>
          {badge !== undefined && (
            <span
              className={cn(
                'px-2 py-0.5 text-xs rounded-full',
                'bg-[var(--surface-brand)] text-[var(--text-inverse)]'
              )}
            >
              {badge}
            </span>
          )}
        </>
      )}
    </a>
  );
}

export interface SidebarSectionProps {
  title?: string;
  children: ReactNode;
  collapsed?: boolean;
  className?: string;
}

export function SidebarSection({
  title,
  children,
  collapsed = false,
  className,
}: SidebarSectionProps) {
  return (
    <div className={cn('py-2', className)}>
      {title && !collapsed && (
        <h3 className="px-3 mb-2 text-xs font-semibold uppercase tracking-wider text-[var(--text-muted)]">
          {title}
        </h3>
      )}
      <div className="space-y-1">{children}</div>
    </div>
  );
}
