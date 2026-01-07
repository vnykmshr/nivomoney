import type { ReactNode } from 'react';
import { cn } from '../../lib/utils';
import { Logo, LogoWithText } from '../Logo';

export interface HeaderProps {
  logo?: ReactNode;
  showLogoText?: boolean;
  nav?: ReactNode;
  actions?: ReactNode;
  className?: string;
  sticky?: boolean;
}

export function Header({
  logo,
  showLogoText = true,
  nav,
  actions,
  className,
  sticky = true,
}: HeaderProps) {
  return (
    <header
      className={cn(
        'h-16 px-4 md:px-6',
        'bg-[var(--surface-card)]',
        'border-b border-[var(--border-subtle)]',
        'flex items-center justify-between',
        sticky && 'sticky top-0 z-40',
        className
      )}
    >
      <div className="flex items-center gap-6">
        {logo || (showLogoText ? <LogoWithText /> : <Logo />)}
        {nav && <nav className="hidden md:flex items-center gap-1">{nav}</nav>}
      </div>

      {actions && <div className="flex items-center gap-2">{actions}</div>}
    </header>
  );
}

export interface NavLinkProps {
  href: string;
  active?: boolean;
  children: ReactNode;
  className?: string;
}

export function NavLink({ href, active, children, className }: NavLinkProps) {
  return (
    <a
      href={href}
      aria-current={active ? 'page' : undefined}
      className={cn(
        'px-3 py-2 rounded-[var(--radius-lg)] text-sm font-medium',
        'transition-colors',
        active
          ? 'bg-[var(--surface-brand-subtle)] text-[var(--interactive-primary)]'
          : 'text-[var(--text-secondary)] hover:text-[var(--text-primary)] hover:bg-[var(--interactive-secondary)]',
        className
      )}
    >
      {children}
    </a>
  );
}

export interface UserMenuProps {
  name: string;
  email?: string;
  avatar?: string;
  onLogout?: () => void;
  className?: string;
}

export function UserMenu({ name, email, avatar, onLogout, className }: UserMenuProps) {
  const initials = name
    .split(' ')
    .map((n) => n[0])
    .join('')
    .toUpperCase()
    .slice(0, 2);

  return (
    <div className={cn('flex items-center gap-3', className)}>
      <div className="text-right hidden sm:block">
        <p className="text-sm font-medium text-[var(--text-primary)]">{name}</p>
        {email && (
          <p className="text-xs text-[var(--text-muted)]">{email}</p>
        )}
      </div>
      <button
        onClick={onLogout}
        className={cn(
          'w-10 h-10 rounded-full',
          'flex items-center justify-center',
          'text-sm font-medium',
          avatar
            ? ''
            : 'bg-[var(--surface-brand)] text-[var(--text-inverse)]',
          'hover:ring-2 hover:ring-[var(--interactive-primary)] hover:ring-offset-2',
          'transition-shadow'
        )}
        title={onLogout ? 'Logout' : name}
      >
        {avatar ? (
          <img
            src={avatar}
            alt={name}
            className="w-full h-full rounded-full object-cover"
          />
        ) : (
          initials
        )}
      </button>
    </div>
  );
}
