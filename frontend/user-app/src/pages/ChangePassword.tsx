import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { api } from '../lib/api';
import { AppLayout } from '../components';
import {
  Alert,
  Button,
  Card,
  FormField,
  Input,
} from '@nivo/shared';

/**
 * Change Password Page
 *
 * User can change their password by providing:
 * - Current password (for verification)
 * - New password
 * - Confirm new password
 */
export function ChangePassword() {
  const navigate = useNavigate();
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);

  const [formData, setFormData] = useState({
    currentPassword: '',
    newPassword: '',
    confirmPassword: '',
  });

  const [errors, setErrors] = useState<Record<string, string>>({});

  const validateForm = (): boolean => {
    const newErrors: Record<string, string> = {};

    if (!formData.currentPassword) {
      newErrors.currentPassword = 'Current password is required';
    }

    if (!formData.newPassword) {
      newErrors.newPassword = 'New password is required';
    } else if (formData.newPassword.length < 8) {
      newErrors.newPassword = 'Password must be at least 8 characters';
    } else if (!/(?=.*[a-z])/.test(formData.newPassword)) {
      newErrors.newPassword = 'Password must contain at least one lowercase letter';
    } else if (!/(?=.*[A-Z])/.test(formData.newPassword)) {
      newErrors.newPassword = 'Password must contain at least one uppercase letter';
    } else if (!/(?=.*\d)/.test(formData.newPassword)) {
      newErrors.newPassword = 'Password must contain at least one number';
    }

    if (formData.newPassword === formData.currentPassword) {
      newErrors.newPassword = 'New password must be different from current password';
    }

    if (!formData.confirmPassword) {
      newErrors.confirmPassword = 'Please confirm your new password';
    } else if (formData.newPassword !== formData.confirmPassword) {
      newErrors.confirmPassword = 'Passwords do not match';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleInputChange = (field: string, value: string) => {
    setFormData({
      ...formData,
      [field]: value,
    });

    // Clear error for this field
    if (errors[field]) {
      setErrors({ ...errors, [field]: '' });
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!validateForm()) {
      return;
    }

    setIsSubmitting(true);
    setError(null);

    try {
      // Change password via API
      await api.changePassword({
        current_password: formData.currentPassword,
        new_password: formData.newPassword,
      });

      setSuccess(true);

      // Redirect to profile after 2 seconds
      setTimeout(() => {
        navigate('/profile');
      }, 2000);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to change password');
    } finally {
      setIsSubmitting(false);
    }
  };

  const getPasswordStrength = (password: string): { score: number; label: string; color: string } => {
    let score = 0;
    if (password.length >= 8) score++;
    if (password.length >= 12) score++;
    if (/[a-z]/.test(password) && /[A-Z]/.test(password)) score++;
    if (/\d/.test(password)) score++;
    if (/[^a-zA-Z0-9]/.test(password)) score++;

    if (score <= 2) return { score, label: 'Weak', color: 'bg-[var(--color-error-500)]' };
    if (score === 3) return { score, label: 'Fair', color: 'bg-[var(--color-warning-500)]' };
    if (score === 4) return { score, label: 'Good', color: 'bg-[var(--interactive-primary)]' };
    return { score, label: 'Strong', color: 'bg-[var(--color-success-500)]' };
  };

  const passwordStrength = formData.newPassword ? getPasswordStrength(formData.newPassword) : null;

  if (success) {
    return (
      <AppLayout title="Password Changed">
        <div className="max-w-md mx-auto px-4 py-6 flex items-center justify-center min-h-[60vh]">
          <Card className="w-full text-center py-8">
            <div className="text-6xl mb-4">✅</div>
            <h2 className="text-2xl font-bold text-[var(--text-primary)] mb-2">Password Changed Successfully!</h2>
            <p className="text-[var(--text-secondary)] mb-4">
              Your password has been updated. You can now login with your new password.
            </p>
            <p className="text-sm text-[var(--text-muted)]">Redirecting to profile...</p>
          </Card>
        </div>
      </AppLayout>
    );
  }

  return (
    <AppLayout title="Change Password" showBack>
      <div className="max-w-md mx-auto px-4 py-6 space-y-6">
        {/* Page Description */}
        <p className="text-[var(--text-secondary)]">Update your password to keep your account secure</p>

        {/* Error Alert */}
        {error && (
          <Alert variant="error" onDismiss={() => setError(null)}>
            {error}
          </Alert>
        )}

        {/* Security Tips */}
        <Alert variant="info">
          <h3 className="text-sm font-semibold mb-2">Password Requirements:</h3>
          <ul className="text-sm space-y-1">
            <li>• At least 8 characters long</li>
            <li>• Contains uppercase and lowercase letters</li>
            <li>• Contains at least one number</li>
            <li>• Different from your current password</li>
          </ul>
        </Alert>

        {/* Change Password Form */}
        <Card padding="lg" className="space-y-6">
          <form onSubmit={handleSubmit} className="space-y-6">
            {/* Current Password */}
            <FormField label="Current Password" htmlFor="currentPassword" error={errors.currentPassword} required>
              <Input
                type="password"
                id="currentPassword"
                value={formData.currentPassword}
                onChange={e => handleInputChange('currentPassword', e.target.value)}
                disabled={isSubmitting}
                error={!!errors.currentPassword}
              />
            </FormField>

            {/* New Password */}
            <div className="space-y-1.5">
              <FormField label="New Password" htmlFor="newPassword" error={errors.newPassword} required>
                <Input
                  type="password"
                  id="newPassword"
                  value={formData.newPassword}
                  onChange={e => handleInputChange('newPassword', e.target.value)}
                  disabled={isSubmitting}
                  error={!!errors.newPassword}
                />
              </FormField>

              {/* Password Strength Indicator */}
              {passwordStrength && (
                <div className="mt-2">
                  <div className="flex items-center justify-between mb-1">
                    <span className="text-xs text-[var(--text-secondary)]">Password Strength</span>
                    <span className="text-xs font-semibold text-[var(--text-primary)]">{passwordStrength.label}</span>
                  </div>
                  <div className="h-2 bg-[var(--surface-muted)] rounded-full overflow-hidden">
                    <div
                      className={`h-full ${passwordStrength.color} transition-all duration-300`}
                      style={{ width: `${(passwordStrength.score / 5) * 100}%` }}
                    />
                  </div>
                </div>
              )}
            </div>

            {/* Confirm New Password */}
            <FormField label="Confirm New Password" htmlFor="confirmPassword" error={errors.confirmPassword} required>
              <Input
                type="password"
                id="confirmPassword"
                value={formData.confirmPassword}
                onChange={e => handleInputChange('confirmPassword', e.target.value)}
                disabled={isSubmitting}
                error={!!errors.confirmPassword}
              />
            </FormField>

            {/* Submit Button */}
            <div className="flex gap-4 pt-4">
              <Button
                type="button"
                variant="secondary"
                onClick={() => navigate('/profile')}
                disabled={isSubmitting}
                className="flex-1"
              >
                Cancel
              </Button>
              <Button
                type="submit"
                loading={isSubmitting}
                className="flex-1"
              >
                Change Password
              </Button>
            </div>
          </form>
        </Card>

        {/* Security Notice */}
        <Card className="bg-[var(--surface-muted)]">
          <div className="flex items-start space-x-3">
            <svg className="w-5 h-5 text-[var(--text-secondary)] mt-0.5" fill="currentColor" viewBox="0 0 20 20">
              <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clipRule="evenodd" />
            </svg>
            <div className="text-sm text-[var(--text-primary)]">
              <p className="font-semibold mb-1">Security Tip</p>
              <p className="text-[var(--text-secondary)]">After changing your password, you'll remain logged in on this device. You may need to log in again on other devices.</p>
            </div>
          </div>
        </Card>
      </div>
    </AppLayout>
  );
}
