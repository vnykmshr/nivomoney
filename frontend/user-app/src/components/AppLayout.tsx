import { ReactNode } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { useAuthStore } from '../stores/authStore';
import {
  LogoWithText,
  Button,
  Avatar,
  BottomNav,
  BottomNavItem,
} from '../../../shared/components';
import { cn } from '../../../shared/lib/utils';

export interface AppLayoutProps {
  children: ReactNode;
  title?: string;
  showBack?: boolean;
  actions?: ReactNode;
}

const navItems = [
  {
    href: '/dashboard',
    label: 'Home',
    icon: (
      <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6" />
      </svg>
    ),
  },
  {
    href: '/send',
    label: 'Send',
    icon: (
      <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8" />
      </svg>
    ),
  },
  {
    href: '/add-money',
    label: 'Add',
    icon: (
      <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
      </svg>
    ),
  },
  {
    href: '/profile',
    label: 'Profile',
    icon: (
      <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
      </svg>
    ),
  },
];

export function AppLayout({ children, title, showBack, actions }: AppLayoutProps) {
  const { user, logout } = useAuthStore();
  const location = useLocation();
  const navigate = useNavigate();

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  return (
    <div className="min-h-screen bg-[var(--surface-page)]">
      {/* Desktop Header */}
      <header className="sticky top-0 z-40 bg-[var(--surface-card)] border-b border-[var(--border-subtle)]">
        <div className="max-w-5xl mx-auto px-4 h-16 flex items-center justify-between">
          {/* Left side */}
          <div className="flex items-center gap-4">
            {showBack ? (
              <button
                onClick={() => navigate(-1)}
                className={cn(
                  'p-2 -ml-2 rounded-lg',
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
            ) : (
              <LogoWithText className="hidden md:flex" />
            )}
            {title && (
              <h1 className="text-lg font-semibold text-[var(--text-primary)]">
                {title}
              </h1>
            )}
          </div>

          {/* Mobile Logo (centered) */}
          {!title && !showBack && (
            <LogoWithText className="md:hidden absolute left-1/2 -translate-x-1/2" />
          )}

          {/* Right side */}
          <div className="flex items-center gap-3">
            {actions}

            {/* Desktop navigation */}
            <nav className="hidden md:flex items-center gap-1">
              {navItems.slice(0, -1).map(item => (
                <a
                  key={item.href}
                  href={item.href}
                  aria-current={location.pathname === item.href ? 'page' : undefined}
                  className={cn(
                    'px-3 py-2 rounded-lg text-sm font-medium',
                    'transition-colors',
                    location.pathname === item.href
                      ? 'bg-[var(--surface-brand-subtle)] text-[var(--interactive-primary)]'
                      : 'text-[var(--text-secondary)] hover:text-[var(--text-primary)] hover:bg-[var(--interactive-secondary)]'
                  )}
                >
                  {item.label}
                </a>
              ))}
            </nav>

            {/* User menu */}
            <div className="flex items-center gap-2">
              <Avatar
                name={user?.full_name || 'User'}
                size="sm"
                className="cursor-pointer"
                onClick={() => navigate('/profile')}
              />
              <Button
                variant="ghost"
                size="sm"
                onClick={handleLogout}
                className="hidden md:inline-flex"
              >
                Logout
              </Button>
            </div>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="pb-20 md:pb-8">
        {children}
      </main>

      {/* Mobile Bottom Navigation */}
      <BottomNav>
        {navItems.map(item => (
          <BottomNavItem
            key={item.href}
            href={item.href}
            icon={item.icon}
            label={item.label}
            active={location.pathname === item.href}
          />
        ))}
      </BottomNav>
    </div>
  );
}
