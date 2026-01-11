/**
 * Skip Link - Accessibility component for keyboard navigation (WCAG 2.4.1)
 *
 * Allows users to skip repetitive navigation and go directly to main content.
 * Only visible when focused (Tab key).
 */
export interface SkipLinkProps {
  href?: string;
  children?: React.ReactNode;
}

export function SkipLink({
  href = '#main-content',
  children = 'Skip to main content'
}: SkipLinkProps) {
  return (
    <a
      href={href}
      className="sr-only focus:not-sr-only focus:absolute focus:top-4 focus:left-4 focus:z-50 focus:px-4 focus:py-2 focus:bg-[var(--surface-card)] focus:text-[var(--text-primary)] focus:border focus:border-[var(--border-default)] focus:rounded-[var(--radius-button)] focus:[box-shadow:var(--focus-ring)]"
    >
      {children}
    </a>
  );
}
