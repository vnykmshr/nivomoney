import { useState, type FormEvent } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { useAdminAuthStore } from '../stores/adminAuthStore';
import {
  Card,
  Button,
  Input,
  FormField,
  Alert,
  Badge,
  LogoWithText,
} from '@nivo/shared';

export function AdminLogin() {
  const navigate = useNavigate();
  const location = useLocation();
  const { login, error, clearError } = useAdminAuthStore();

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
        {/* Admin Branding */}
        <div className="text-center mb-6">
          <LogoWithText size="lg" className="justify-center mb-4" />
          <Badge className="admin-badge px-4 py-2 text-sm">
            Admin Portal
          </Badge>
          <p className="mt-3 text-[var(--text-secondary)]">
            Secure administrative access
          </p>
        </div>

        {/* Login Card */}
        <Card padding="lg">
          {/* Security Notice */}
          <Alert variant="warning" className="mb-6">
            <strong>Security Notice:</strong> Admin sessions expire after 2 hours of inactivity.
            All admin actions are logged and monitored.
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
                placeholder="admin@nivomoney.com"
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
                placeholder="••••••••"
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
              Admin Login
            </Button>
          </form>

          {/* Security Footer */}
          <div className="mt-6 pt-6 border-t border-[var(--border-subtle)]">
            <div className="flex items-start gap-3 text-xs text-[var(--text-muted)]">
              <svg className="w-4 h-4 mt-0.5 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
              </svg>
              <div>
                <p className="font-medium mb-1 text-[var(--text-secondary)]">Security Features:</p>
                <ul className="list-disc list-inside space-y-1">
                  <li>2-hour session timeout</li>
                  <li>Activity-based auto-logout</li>
                  <li>Audit logging enabled</li>
                  <li>VPN required in production</li>
                </ul>
              </div>
            </div>
          </div>
        </Card>

        {/* Help Text */}
        <p className="mt-6 text-center text-sm text-[var(--text-muted)]">
          Admin credentials required. Contact IT support for access.
        </p>
      </div>
    </div>
  );
}
