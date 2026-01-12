import { useState } from 'react';
import type { ReactNode } from 'react';
import { useLocation, useNavigate, Link } from 'react-router-dom';
import { useAdminAuthStore } from '../stores/adminAuthStore';
import { cn } from '../../../shared/lib/utils';
import { Button, Badge, Breadcrumbs, type BreadcrumbItem } from '../../../shared/components';

export interface AdminLayoutProps {
  children: ReactNode;
  title?: string;
  breadcrumbs?: BreadcrumbItem[];
}

const navItems = [
  {
    href: '/',
    label: 'Dashboard',
    icon: (
      <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6" />
      </svg>
    ),
  },
  {
    href: '/kyc',
    label: 'KYC Reviews',
    icon: (
      <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
      </svg>
    ),
  },
  {
    href: '/users',
    label: 'Users',
    icon: (
      <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
      </svg>
    ),
  },
  {
    href: '/transactions',
    label: 'Transactions',
    icon: (
      <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
      </svg>
    ),
  },
];

// Check if a nav item is active (exact match for root, prefix match for others)
function isNavItemActive(pathname: string, href: string): boolean {
  if (href === '/') {
    return pathname === '/';
  }
  return pathname === href || pathname.startsWith(href + '/');
}

export function AdminLayout({ children, title, breadcrumbs }: AdminLayoutProps) {
  const { user, logout } = useAdminAuthStore();
  const location = useLocation();
  const navigate = useNavigate();
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false);
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);

  const handleLogout = async () => {
    await logout();
    navigate('/login');
  };

  return (
    <div className="min-h-screen bg-[var(--surface-page)] flex">
      {/* Sidebar */}
      <aside
        className={cn(
          'hidden md:flex flex-col',
          'bg-[var(--surface-card)]',
          'border-r border-[var(--border-subtle)]',
          'transition-[width] duration-200',
          sidebarCollapsed ? 'w-16' : 'w-64'
        )}
      >
        {/* Logo */}
        <div className="h-16 flex items-center justify-between px-4 border-b border-[var(--border-subtle)]">
          {!sidebarCollapsed && (
            <div className="flex items-center gap-2">
              <div className="w-8 h-8 rounded-lg bg-[var(--surface-brand)] flex items-center justify-center">
                <span className="text-[var(--text-inverse)] font-bold text-sm">N</span>
              </div>
              <span className="font-semibold text-[var(--text-primary)]">Admin</span>
            </div>
          )}
          <button
            onClick={() => setSidebarCollapsed(!sidebarCollapsed)}
            className={cn(
              'p-1.5 rounded-lg',
              'text-[var(--text-muted)] hover:text-[var(--text-primary)]',
              'hover:bg-[var(--interactive-secondary)]',
              'transition-colors',
              sidebarCollapsed && 'mx-auto'
            )}
            aria-label={sidebarCollapsed ? 'Expand sidebar' : 'Collapse sidebar'}
          >
            <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
              {sidebarCollapsed ? (
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 5l7 7-7 7M5 5l7 7-7 7" />
              ) : (
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11 19l-7-7 7-7m8 14l-7-7 7-7" />
              )}
            </svg>
          </button>
        </div>

        {/* Navigation */}
        <nav className="flex-1 p-3 space-y-1">
          {navItems.map(item => (
            <Link
              key={item.href}
              to={item.href}
              aria-current={isNavItemActive(location.pathname, item.href) ? 'page' : undefined}
              className={cn(
                'flex items-center gap-3 px-3 py-2.5',
                'rounded-lg',
                'text-sm font-medium',
                'transition-colors',
                isNavItemActive(location.pathname, item.href)
                  ? 'bg-[var(--surface-brand-subtle)] text-[var(--interactive-primary)]'
                  : 'text-[var(--text-secondary)] hover:text-[var(--text-primary)] hover:bg-[var(--interactive-secondary)]',
                sidebarCollapsed && 'justify-center px-2'
              )}
              title={sidebarCollapsed ? item.label : undefined}
            >
              <span className="w-5 h-5 flex-shrink-0">{item.icon}</span>
              {!sidebarCollapsed && <span className="flex-1">{item.label}</span>}
            </Link>
          ))}
        </nav>

        {/* User section */}
        <div className="p-3 border-t border-[var(--border-subtle)]">
          {!sidebarCollapsed ? (
            <div className="flex items-center gap-3 px-3 py-2">
              <div className="w-8 h-8 rounded-full bg-[var(--color-accent-100)] flex items-center justify-center">
                <span className="text-sm font-medium text-[var(--color-accent-700)]">
                  {(user?.full_name || 'A').charAt(0).toUpperCase()}
                </span>
              </div>
              <div className="flex-1 min-w-0">
                <p className="text-sm font-medium text-[var(--text-primary)] truncate">
                  {user?.full_name}
                </p>
                <p className="text-xs text-[var(--text-muted)]">Admin</p>
              </div>
            </div>
          ) : (
            <div className="flex justify-center">
              <div className="w-8 h-8 rounded-full bg-[var(--color-accent-100)] flex items-center justify-center">
                <span className="text-sm font-medium text-[var(--color-accent-700)]">
                  {(user?.full_name || 'A').charAt(0).toUpperCase()}
                </span>
              </div>
            </div>
          )}
        </div>
      </aside>

      {/* Main content */}
      <div className="flex-1 flex flex-col">
        {/* Header */}
        <header className="h-16 bg-[var(--surface-card)] border-b border-[var(--border-subtle)] flex items-center justify-between px-6">
          <div className="flex items-center gap-4">
            {/* Mobile menu button */}
            <button
              onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
              className="md:hidden p-2 rounded-lg text-[var(--text-secondary)] hover:bg-[var(--interactive-secondary)]"
              aria-label={mobileMenuOpen ? 'Close menu' : 'Open menu'}
              aria-expanded={mobileMenuOpen}
            >
              <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                {mobileMenuOpen ? (
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                ) : (
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16M4 18h16" />
                )}
              </svg>
            </button>

            {title && (
              <h1 className="text-xl font-semibold text-[var(--text-primary)]">
                {title}
              </h1>
            )}
          </div>

          <div className="flex items-center gap-4">
            <Badge className="admin-badge">Admin</Badge>
            <span className="text-sm text-[var(--text-secondary)] hidden sm:block">
              {user?.full_name}
            </span>
            <Button variant="secondary" size="sm" onClick={handleLogout}>
              Logout
            </Button>
          </div>
        </header>

        {/* Page content */}
        <main className="flex-1 p-6">
          {breadcrumbs && breadcrumbs.length > 0 && (
            <Breadcrumbs items={breadcrumbs} />
          )}
          <div className="animate-[fadeIn_0.2s_ease-out]">
            {children}
          </div>
        </main>
      </div>

      {/* Mobile menu overlay */}
      {mobileMenuOpen && (
        <div className="md:hidden fixed inset-0 z-50">
          {/* Backdrop */}
          <div
            className="fixed inset-0 bg-[var(--surface-overlay)]"
            onClick={() => setMobileMenuOpen(false)}
            aria-hidden="true"
          />

          {/* Menu drawer */}
          <div className="fixed inset-y-0 left-0 w-64 bg-[var(--surface-card)] shadow-xl">
            {/* Logo */}
            <div className="h-16 flex items-center justify-between px-4 border-b border-[var(--border-subtle)]">
              <div className="flex items-center gap-2">
                <div className="w-8 h-8 rounded-lg bg-[var(--surface-brand)] flex items-center justify-center">
                  <span className="text-[var(--text-inverse)] font-bold text-sm">N</span>
                </div>
                <span className="font-semibold text-[var(--text-primary)]">Admin</span>
              </div>
              <button
                onClick={() => setMobileMenuOpen(false)}
                className="p-1.5 rounded-lg text-[var(--text-muted)] hover:text-[var(--text-primary)] hover:bg-[var(--interactive-secondary)]"
                aria-label="Close menu"
              >
                <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>

            {/* Navigation */}
            <nav className="flex-1 p-3 space-y-1">
              {navItems.map(item => (
                <Link
                  key={item.href}
                  to={item.href}
                  onClick={() => setMobileMenuOpen(false)}
                  aria-current={isNavItemActive(location.pathname, item.href) ? 'page' : undefined}
                  className={cn(
                    'flex items-center gap-3 px-3 py-2.5',
                    'rounded-lg',
                    'text-sm font-medium',
                    'transition-colors',
                    isNavItemActive(location.pathname, item.href)
                      ? 'bg-[var(--surface-brand-subtle)] text-[var(--interactive-primary)]'
                      : 'text-[var(--text-secondary)] hover:text-[var(--text-primary)] hover:bg-[var(--interactive-secondary)]'
                  )}
                >
                  <span className="w-5 h-5 flex-shrink-0">{item.icon}</span>
                  <span className="flex-1">{item.label}</span>
                </Link>
              ))}
            </nav>

            {/* User section */}
            <div className="p-3 border-t border-[var(--border-subtle)]">
              <div className="flex items-center gap-3 px-3 py-2">
                <div className="w-8 h-8 rounded-full bg-[var(--color-accent-100)] flex items-center justify-center">
                  <span className="text-sm font-medium text-[var(--color-accent-700)]">
                    {(user?.full_name || 'A').charAt(0).toUpperCase()}
                  </span>
                </div>
                <div className="flex-1 min-w-0">
                  <p className="text-sm font-medium text-[var(--text-primary)] truncate">
                    {user?.full_name}
                  </p>
                  <p className="text-xs text-[var(--text-muted)]">Admin</p>
                </div>
              </div>
              <Button
                variant="secondary"
                size="sm"
                onClick={() => {
                  setMobileMenuOpen(false);
                  handleLogout();
                }}
                className="w-full mt-2"
              >
                Logout
              </Button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
