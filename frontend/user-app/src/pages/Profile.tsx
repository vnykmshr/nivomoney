import { useState, useEffect, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuthStore } from '../stores/authStore';
import { api } from '../lib/api';
import { formatDateShort } from '@nivo/shared';
import { AppLayout } from '../components';
import {
  Card,
  CardTitle,
  Button,
  Input,
  FormField,
  Alert,
  Avatar,
  Badge,
  Skeleton,
} from '@nivo/shared';
import type { User } from '@nivo/shared';

export function Profile() {
  const navigate = useNavigate();
  const { user: authUser, updateUser } = useAuthStore();
  const [user, setUserLocal] = useState<User | null>(authUser);
  const [isEditing, setIsEditing] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  const [formData, setFormData] = useState({
    full_name: authUser?.full_name || '',
    email: authUser?.email || '',
    phone: authUser?.phone || '',
  });

  const [errors, setErrors] = useState<Record<string, string>>({});

  const fetchProfile = useCallback(async () => {
    try {
      setIsLoading(true);
      const profileData = await api.getProfile();
      setUserLocal(profileData);
      updateUser(profileData);
      setFormData({
        full_name: profileData.full_name,
        email: profileData.email,
        phone: profileData.phone,
      });
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load profile');
    } finally {
      setIsLoading(false);
    }
  }, [updateUser]);

  useEffect(() => {
    fetchProfile();
  }, [fetchProfile]);

  const validateForm = (): boolean => {
    const newErrors: Record<string, string> = {};

    if (!formData.full_name.trim()) {
      newErrors.full_name = 'Full name is required';
    } else if (formData.full_name.trim().length < 2) {
      newErrors.full_name = 'Full name must be at least 2 characters';
    }

    if (!formData.email.trim()) {
      newErrors.email = 'Email is required';
    } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(formData.email)) {
      newErrors.email = 'Invalid email format';
    }

    if (!formData.phone.trim()) {
      newErrors.phone = 'Phone number is required';
    } else if (!/^(\+91)?[6-9]\d{9}$/.test(formData.phone.replace(/\s/g, ''))) {
      newErrors.phone = 'Invalid phone number (10 digits starting with 6-9)';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleInputChange = (field: string, value: string) => {
    setFormData({ ...formData, [field]: value });
    if (errors[field]) {
      setErrors({ ...errors, [field]: '' });
    }
  };

  const handleEdit = () => {
    setIsEditing(true);
    setError(null);
    setSuccess(null);
  };

  const handleCancel = () => {
    setIsEditing(false);
    setError(null);
    setSuccess(null);
    if (user) {
      setFormData({
        full_name: user.full_name,
        email: user.email,
        phone: user.phone,
      });
    }
    setErrors({});
  };

  const handleSave = async () => {
    if (!validateForm()) return;

    setIsSaving(true);
    setError(null);
    setSuccess(null);

    try {
      const updatedUser = await api.updateProfile(formData);
      setUserLocal(updatedUser);
      updateUser(updatedUser);
      setSuccess('Profile updated successfully!');
      setIsEditing(false);
      setTimeout(() => setSuccess(null), 3000);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to update profile');
    } finally {
      setIsSaving(false);
    }
  };

  if (isLoading || !user) {
    return (
      <AppLayout title="Profile">
        <div className="max-w-2xl mx-auto px-4 py-6">
          <Card>
            <div className="space-y-4">
              <div className="flex items-center gap-4">
                <Skeleton className="w-16 h-16 rounded-full" />
                <div className="space-y-2">
                  <Skeleton className="h-6 w-32" />
                  <Skeleton className="h-4 w-24" />
                </div>
              </div>
              <Skeleton className="h-10" />
              <Skeleton className="h-10" />
              <Skeleton className="h-10" />
            </div>
          </Card>
        </div>
      </AppLayout>
    );
  }

  return (
    <AppLayout title="Profile">
      <div className="max-w-2xl mx-auto px-4 py-6 space-y-6">
        {success && <Alert variant="success">{success}</Alert>}
        {error && <Alert variant="error">{error}</Alert>}

        {/* Profile Card */}
        <Card>
          {/* Profile Header */}
          <div className="flex items-center justify-between mb-6 pb-6 border-b border-[var(--border-subtle)]">
            <div className="flex items-center gap-4">
              <Avatar name={user.full_name} size="lg" />
              <div>
                <h3 className="text-xl font-semibold text-[var(--text-primary)]">
                  {user.full_name}
                </h3>
                <p className="text-sm text-[var(--text-muted)]">
                  Member since {formatDateShort(user.created_at)}
                </p>
              </div>
            </div>
            {!isEditing && (
              <Button onClick={handleEdit}>Edit Profile</Button>
            )}
          </div>

          {/* Profile Form */}
          <div className="space-y-5">
            {/* Account Status */}
            <div>
              <label className="block text-sm font-medium text-[var(--text-primary)] mb-2">
                Account Status
              </label>
              <div className="flex items-center gap-2">
                <Badge
                  variant={user.status === 'active' ? 'success' : user.status === 'pending' ? 'warning' : 'error'}
                >
                  {user.status.toUpperCase()}
                </Badge>
                {user.kyc && (
                  <Badge
                    variant={user.kyc.status === 'verified' ? 'success' : user.kyc.status === 'pending' ? 'warning' : 'error'}
                  >
                    KYC: {user.kyc.status.toUpperCase()}
                  </Badge>
                )}
              </div>
            </div>

            {/* Full Name */}
            <FormField
              label="Full Name"
              htmlFor="full_name"
              error={errors.full_name}
              required={isEditing}
            >
              <Input
                id="full_name"
                value={formData.full_name}
                onChange={e => handleInputChange('full_name', e.target.value)}
                disabled={!isEditing || isSaving}
                error={!!errors.full_name}
              />
            </FormField>

            {/* Email */}
            <FormField
              label="Email Address"
              htmlFor="email"
              error={errors.email}
              hint={isEditing ? 'Changing email will require verification (future feature)' : undefined}
              required={isEditing}
            >
              <Input
                id="email"
                type="email"
                value={formData.email}
                onChange={e => handleInputChange('email', e.target.value)}
                disabled={!isEditing || isSaving}
                error={!!errors.email}
              />
            </FormField>

            {/* Phone */}
            <FormField
              label="Phone Number"
              htmlFor="phone"
              error={errors.phone}
              hint={isEditing ? 'Changing phone will require verification (future feature)' : undefined}
              required={isEditing}
            >
              <Input
                id="phone"
                type="tel"
                value={formData.phone}
                onChange={e => handleInputChange('phone', e.target.value)}
                disabled={!isEditing || isSaving}
                error={!!errors.phone}
              />
            </FormField>

            {/* User ID */}
            <div>
              <label className="block text-sm font-medium text-[var(--text-primary)] mb-2">
                User ID
              </label>
              <div className="p-3 bg-[var(--surface-page)] rounded-lg">
                <code className="text-sm font-mono text-[var(--text-secondary)]">{user.id}</code>
              </div>
            </div>

            {/* Last Updated */}
            <div>
              <label className="block text-sm font-medium text-[var(--text-primary)] mb-2">
                Last Updated
              </label>
              <p className="text-sm text-[var(--text-secondary)]">
                {formatDateShort(user.updated_at)}
              </p>
            </div>
          </div>

          {/* Action Buttons */}
          {isEditing && (
            <div className="flex gap-3 mt-6 pt-6 border-t border-[var(--border-subtle)]">
              <Button
                variant="secondary"
                className="flex-1"
                onClick={handleCancel}
                disabled={isSaving}
              >
                Cancel
              </Button>
              <Button
                className="flex-1"
                onClick={handleSave}
                loading={isSaving}
              >
                Save Changes
              </Button>
            </div>
          )}
        </Card>

        {/* Account Actions */}
        <Card>
          <CardTitle className="mb-4">Account Actions</CardTitle>
          <div className="space-y-2">
            <button
              onClick={() => navigate('/change-password')}
              className="w-full flex items-center justify-between p-4 rounded-lg bg-[var(--surface-page)] hover:bg-[var(--interactive-secondary)] transition-colors"
            >
              <div className="flex items-center gap-3">
                <svg className="w-5 h-5 text-[var(--text-secondary)]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
                </svg>
                <span className="font-medium text-[var(--text-primary)]">Change Password</span>
              </div>
              <svg className="w-5 h-5 text-[var(--text-muted)]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
              </svg>
            </button>

            <button
              onClick={() => navigate('/kyc')}
              className="w-full flex items-center justify-between p-4 rounded-lg bg-[var(--surface-page)] hover:bg-[var(--interactive-secondary)] transition-colors"
            >
              <div className="flex items-center gap-3">
                <svg className="w-5 h-5 text-[var(--text-secondary)]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                </svg>
                <span className="font-medium text-[var(--text-primary)]">KYC Verification</span>
              </div>
              <svg className="w-5 h-5 text-[var(--text-muted)]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
              </svg>
            </button>
          </div>
        </Card>
      </div>
    </AppLayout>
  );
}
