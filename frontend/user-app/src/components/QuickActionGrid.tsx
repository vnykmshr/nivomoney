import type { ReactNode } from 'react';
import { Link } from 'react-router-dom';
import { cn } from '@nivo/shared';

export interface QuickAction {
  id: string;
  label: string;
  description?: string;
  href: string;
  icon: ReactNode;
  color?: 'primary' | 'success' | 'warning' | 'neutral';
}

export interface QuickActionGridProps {
  actions: QuickAction[];
  className?: string;
}

const colorVariants = {
  primary: {
    bg: 'bg-[var(--color-primary-50)]',
    icon: 'text-[var(--color-primary-600)]',
    hover: 'hover:bg-[var(--color-primary-100)]',
  },
  success: {
    bg: 'bg-[var(--color-success-50)]',
    icon: 'text-[var(--color-success-600)]',
    hover: 'hover:bg-[var(--color-success-100)]',
  },
  warning: {
    bg: 'bg-[var(--color-warning-50)]',
    icon: 'text-[var(--color-warning-600)]',
    hover: 'hover:bg-[var(--color-warning-100)]',
  },
  neutral: {
    bg: 'bg-[var(--color-neutral-100)]',
    icon: 'text-[var(--color-neutral-600)]',
    hover: 'hover:bg-[var(--color-neutral-200)]',
  },
};

export function QuickActionGrid({ actions, className }: QuickActionGridProps) {
  return (
    <div className={cn('grid grid-cols-2 md:grid-cols-4 gap-3', className)}>
      {actions.map(action => {
        const colors = colorVariants[action.color || 'primary'];

        return (
          <Link
            key={action.id}
            to={action.href}
            className={cn(
              'flex flex-col items-center p-4 rounded-xl',
              'transition-all duration-150',
              'focus:outline-none focus-visible:[box-shadow:var(--focus-ring)]',
              colors.bg,
              colors.hover
            )}
          >
            <div
              className={cn(
                'w-12 h-12 rounded-full flex items-center justify-center mb-3',
                'bg-[var(--surface-card)] shadow-sm',
                colors.icon
              )}
            >
              {action.icon}
            </div>
            <span className="text-sm font-medium text-[var(--text-primary)] text-center">
              {action.label}
            </span>
            {action.description && (
              <span className="text-xs text-[var(--text-muted)] mt-1 text-center">
                {action.description}
              </span>
            )}
          </Link>
        );
      })}
    </div>
  );
}

// Predefined icons for common actions
export const QuickActionIcons = {
  addMoney: (
    <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
    </svg>
  ),
  sendMoney: (
    <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8" />
    </svg>
  ),
  withdraw: (
    <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 9V7a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2m2 4h10a2 2 0 002-2v-6a2 2 0 00-2-2H9a2 2 0 00-2 2v6a2 2 0 002 2zm7-5a2 2 0 11-4 0 2 2 0 014 0z" />
    </svg>
  ),
  beneficiaries: (
    <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
    </svg>
  ),
  history: (
    <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
    </svg>
  ),
  profile: (
    <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
    </svg>
  ),
  settings: (
    <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
    </svg>
  ),
};
