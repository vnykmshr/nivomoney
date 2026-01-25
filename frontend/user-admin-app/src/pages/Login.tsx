import { useState, type FormEvent } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { useAuthStore } from '../stores/authStore';
import {
  Card,
  Button,
  Input,
  FormField,
  Alert,
  Badge,
  LogoWithText,
} from '@nivo/shared';

export function Login() {
  const navigate = useNavigate();
  const location = useLocation();
  const { login, error, clearError } = useAuthStore();

  const [identifier, setIdentifier] = useState('');
  const [password, setPassword] = useState('');
  const [isLoading, setIsLoading] = useState(false);

  const from = (location.state as { from?: { pathname: string } })?.from?.pathname || '/';

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    clearError();

    if (!identifier || !password) {
      return;
    }

    try {
      setIsLoading(true);
      await login(identifier, password);
      navigate(from, { replace: true });
    } catch {
      // Error is handled by the auth store
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-[var(--color-primary-50)] to-[var(--color-neutral-100)] flex items-center justify-center p-4">
      <div className="max-w-md w-full">
        {/* Portal Branding */}
        <div className="text-center mb-6">
          <LogoWithText size="lg" className="justify-center mb-4" />
          <Badge variant="info" className="px-4 py-2 text-sm">
            Verification Portal
          </Badge>
          <p className="mt-3 text-[var(--text-secondary)]">
            View OTP codes for your paired user
          </p>
        </div>

        {/* Login Card */}
        <Card padding="lg">
          {/* Purpose Notice */}
          <Alert variant="info" className="mb-6">
            <strong>Verification Portal:</strong> View OTP codes for your paired user's transactions.
          </Alert>

          {/* Error Message */}
          {error && (
            <Alert
              variant="error"
              className="mb-4"
              onDismiss={clearError}
            >
              {error}
            </Alert>
          )}

          {/* Login Form */}
          <form onSubmit={handleSubmit} className="space-y-5">
            <FormField
              label="Email or Phone"
              htmlFor="identifier"
              required
            >
              <Input
                id="identifier"
                type="text"
                value={identifier}
                onChange={e => setIdentifier(e.target.value)}
                placeholder="your.email@example.com"
                disabled={isLoading}
                autoComplete="username"
              />
            </FormField>

            <FormField
              label="Password"
              htmlFor="password"
              required
            >
              <Input
                id="password"
                type="password"
                value={password}
                onChange={e => setPassword(e.target.value)}
                placeholder="Enter your password"
                disabled={isLoading}
                autoComplete="current-password"
              />
            </FormField>

            <Button
              type="submit"
              disabled={isLoading || !identifier || !password}
              loading={isLoading}
              className="w-full"
              size="lg"
            >
              Login
            </Button>
          </form>

          {/* Security Footer */}
          <div className="mt-6 pt-6 border-t border-[var(--border-subtle)]">
            <div className="flex items-start gap-3 text-xs text-[var(--text-muted)]">
              <svg className="w-4 h-4 mt-0.5 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
              </svg>
              <div>
                <p className="font-medium mb-1 text-[var(--text-secondary)]">Portal Purpose:</p>
                <ul className="list-disc list-inside space-y-1">
                  <li>View verification OTP codes</li>
                  <li>Help your paired user complete transactions</li>
                  <li>Session expires after 1 hour</li>
                </ul>
              </div>
            </div>
          </div>
        </Card>

        {/* Help Text */}
        <p className="mt-6 text-center text-sm text-[var(--text-muted)]">
          User-Admin credentials required. Contact support if you need access.
        </p>
      </div>
    </div>
  );
}
