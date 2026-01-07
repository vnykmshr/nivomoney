import type { ReactNode } from 'react';
import { cn } from '../../lib/utils';

export interface PageLayoutProps {
  children: ReactNode;
  header?: ReactNode;
  footer?: ReactNode;
  sidebar?: ReactNode;
  bottomNav?: ReactNode;
  className?: string;
  contentClassName?: string;
  maxWidth?: 'sm' | 'md' | 'lg' | 'xl' | 'full';
  padding?: boolean;
}

export function PageLayout({
  children,
  header,
  footer,
  sidebar,
  bottomNav,
  className,
  contentClassName,
  maxWidth = 'lg',
  padding = true,
}: PageLayoutProps) {
  return (
    <div
      className={cn(
        'min-h-screen flex flex-col',
        'bg-[var(--surface-page)]',
        className
      )}
    >
      {header}

      <div className="flex-1 flex">
        {sidebar}

        <main
          className={cn(
            'flex-1',
            padding && 'px-4 py-6 md:px-6',
            contentClassName
          )}
        >
          <div
            className={cn(
              'mx-auto w-full',
              maxWidth === 'sm' && 'max-w-sm',
              maxWidth === 'md' && 'max-w-md',
              maxWidth === 'lg' && 'max-w-4xl',
              maxWidth === 'xl' && 'max-w-6xl',
              maxWidth === 'full' && 'max-w-full'
            )}
          >
            {children}
          </div>
        </main>
      </div>

      {footer}
      {bottomNav}
    </div>
  );
}

export interface PageHeaderProps {
  title: string;
  subtitle?: string;
  action?: ReactNode;
  backHref?: string;
  onBack?: () => void;
  className?: string;
}

export function PageHeader({
  title,
  subtitle,
  action,
  backHref,
  onBack,
  className,
}: PageHeaderProps) {
  const BackButton = () => (
    <button
      onClick={onBack}
      className={cn(
        'p-2 -ml-2 rounded-[var(--radius-lg)]',
        'text-[var(--text-secondary)] hover:text-[var(--text-primary)]',
        'hover:bg-[var(--interactive-secondary)]',
        'transition-colors'
      )}
      aria-label="Go back"
    >
      <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
      </svg>
    </button>
  );

  const BackLink = () => (
    <a
      href={backHref}
      className={cn(
        'p-2 -ml-2 rounded-[var(--radius-lg)]',
        'text-[var(--text-secondary)] hover:text-[var(--text-primary)]',
        'hover:bg-[var(--interactive-secondary)]',
        'transition-colors'
      )}
      aria-label="Go back"
    >
      <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
      </svg>
    </a>
  );

  return (
    <div className={cn('flex items-center justify-between mb-6', className)}>
      <div className="flex items-center gap-2">
        {onBack && <BackButton />}
        {backHref && !onBack && <BackLink />}
        <div>
          <h1 className="text-2xl font-bold text-[var(--text-primary)]">
            {title}
          </h1>
          {subtitle && (
            <p className="text-sm text-[var(--text-secondary)] mt-0.5">
              {subtitle}
            </p>
          )}
        </div>
      </div>
      {action && <div>{action}</div>}
    </div>
  );
}
