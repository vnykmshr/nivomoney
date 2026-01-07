import { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { useAuthStore } from '../stores/authStore';
import { isValidEmail, isValidPhone, normalizeIndianPhone } from '../lib/utils';
import {
  LogoWithText,
  Card,
  Button,
  Input,
  FormField,
  Alert,
} from '../../../shared/components';

export function Register() {
  const [formData, setFormData] = useState({
    email: '',
    password: '',
    confirmPassword: '',
    fullName: '',
    phone: '',
  });
  const [errors, setErrors] = useState<Record<string, string>>({});
  const navigate = useNavigate();

  const { register, isLoading, error: authError } = useAuthStore();

  const validate = (): boolean => {
    const newErrors: Record<string, string> = {};

    if (!formData.fullName) {
      newErrors.fullName = 'Full name is required';
    }

    if (!formData.email) {
      newErrors.email = 'Email is required';
    } else if (!isValidEmail(formData.email)) {
      newErrors.email = 'Invalid email format';
    }

    if (!formData.phone) {
      newErrors.phone = 'Phone number is required';
    } else if (!isValidPhone(formData.phone)) {
      newErrors.phone = 'Invalid phone number';
    }

    if (!formData.password) {
      newErrors.password = 'Password is required';
    } else if (formData.password.length < 6) {
      newErrors.password = 'Password must be at least 6 characters';
    }

    if (formData.password !== formData.confirmPassword) {
      newErrors.confirmPassword = 'Passwords do not match';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!validate()) return;

    try {
      const normalizedPhone = normalizeIndianPhone(formData.phone);
      await register(formData.email, formData.password, formData.fullName, normalizedPhone);
      navigate('/dashboard');
    } catch (err) {
      console.error('Registration failed:', err);
    }
  };

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setFormData(prev => ({
      ...prev,
      [e.target.name]: e.target.value,
    }));
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-[var(--surface-page)] py-12 px-4">
      <div className="w-full max-w-md">
        {/* Logo */}
        <div className="text-center mb-8">
          <LogoWithText className="justify-center" />
          <p className="mt-2 text-[var(--text-secondary)]">
            Create your account
          </p>
        </div>

        <Card padding="lg">
          <form onSubmit={handleSubmit} className="space-y-4">
            {authError && (
              <Alert variant="error">
                {authError}
              </Alert>
            )}

            <FormField
              label="Full Name"
              htmlFor="fullName"
              error={errors.fullName}
              required
            >
              <Input
                id="fullName"
                name="fullName"
                type="text"
                value={formData.fullName}
                onChange={handleChange}
                placeholder="John Doe"
                disabled={isLoading}
                error={!!errors.fullName}
                autoComplete="name"
              />
            </FormField>

            <FormField
              label="Email"
              htmlFor="email"
              error={errors.email}
              required
            >
              <Input
                id="email"
                name="email"
                type="email"
                value={formData.email}
                onChange={handleChange}
                placeholder="you@example.com"
                disabled={isLoading}
                error={!!errors.email}
                autoComplete="email"
              />
            </FormField>

            <FormField
              label="Phone Number"
              htmlFor="phone"
              error={errors.phone}
              hint="Enter 10-digit mobile number"
              required
            >
              <Input
                id="phone"
                name="phone"
                type="tel"
                value={formData.phone}
                onChange={handleChange}
                placeholder="9876543210"
                disabled={isLoading}
                error={!!errors.phone}
                autoComplete="tel"
              />
            </FormField>

            <FormField
              label="Password"
              htmlFor="password"
              error={errors.password}
              hint="At least 6 characters"
              required
            >
              <Input
                id="password"
                name="password"
                type="password"
                value={formData.password}
                onChange={handleChange}
                placeholder="Create a password"
                disabled={isLoading}
                error={!!errors.password}
                autoComplete="new-password"
              />
            </FormField>

            <FormField
              label="Confirm Password"
              htmlFor="confirmPassword"
              error={errors.confirmPassword}
              required
            >
              <Input
                id="confirmPassword"
                name="confirmPassword"
                type="password"
                value={formData.confirmPassword}
                onChange={handleChange}
                placeholder="Confirm your password"
                disabled={isLoading}
                error={!!errors.confirmPassword}
                autoComplete="new-password"
              />
            </FormField>

            <Button
              type="submit"
              className="w-full"
              loading={isLoading}
              size="lg"
            >
              Create Account
            </Button>
          </form>

          <div className="mt-6 text-center">
            <p className="text-sm text-[var(--text-secondary)]">
              Already have an account?{' '}
              <Link
                to="/login"
                className="text-[var(--text-link)] hover:text-[var(--text-link-hover)] font-medium"
              >
                Sign in
              </Link>
            </p>
          </div>
        </Card>

        {/* Footer */}
        <p className="mt-8 text-center text-xs text-[var(--text-muted)]">
          By creating an account, you agree to our Terms of Service and Privacy Policy.
        </p>
      </div>
    </div>
  );
}
