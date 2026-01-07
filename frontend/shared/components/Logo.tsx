interface LogoProps {
  className?: string;
  size?: 'sm' | 'md' | 'lg';
  'aria-label'?: string;
}

export function Logo({
  className = '',
  size = 'md',
  'aria-label': ariaLabel = 'Nivo Money',
}: LogoProps) {
  const sizes = {
    sm: 'h-6 w-6',
    md: 'h-8 w-8',
    lg: 'h-12 w-12',
  };

  return (
    <svg
      className={`${sizes[size]} ${className}`}
      viewBox="0 0 40 40"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      role="img"
      aria-label={ariaLabel}
    >
      <circle cx="20" cy="20" r="18" fill="var(--surface-brand)" />
      <path
        d="M12 20L18 26L28 14"
        stroke="var(--text-inverse)"
        strokeWidth="3"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}

export function LogoWithText({ className = '' }: { className?: string }) {
  return (
    <div className={`flex items-center gap-2 ${className}`}>
      <Logo size="md" />
      <span
        className="font-semibold text-xl"
        style={{ color: 'var(--text-primary)' }}
      >
        Nivo Money
      </span>
    </div>
  );
}
