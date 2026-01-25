import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { api } from '../lib/api';
import type { UpdateKYCRequest } from '@nivo/shared';
import { AppLayout } from '../components';
import {
  Alert,
  Button,
  Card,
  FormField,
  Input,
} from '@nivo/shared';

/**
 * KYC Submission Form
 *
 * IMPORTANT: This form does NOT upload actual documents.
 * It collects document numbers (PAN, Aadhaar) as text for validation.
 *
 * Admin Workflow Pattern:
 * 1. User submits KYC data (text only, no files)
 * 2. Backend creates notification for admin
 * 3. Admin reviews from notification panel in Admin Dashboard
 * 4. Admin approves/rejects
 * 5. User receives status update
 *
 * See: /ADMIN_WORKFLOW_PATTERN.md
 */
export function KYC() {
  const navigate = useNavigate();
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);

  const [formData, setFormData] = useState<UpdateKYCRequest>({
    pan: '',
    aadhaar: '',
    date_of_birth: '',
    address: {
      street: '',
      city: '',
      state: '',
      pin: '',
      country: 'IN',
    },
  });

  const [errors, setErrors] = useState<Record<string, string>>({});

  const validatePAN = (pan: string): boolean => {
    // PAN format: ABCDE1234F (5 letters, 4 digits, 1 letter)
    const panRegex = /^[A-Z]{5}[0-9]{4}[A-Z]{1}$/;
    return panRegex.test(pan.toUpperCase());
  };

  const validateAadhaar = (aadhaar: string): boolean => {
    // Aadhaar: 12 digits
    const aadhaarRegex = /^[0-9]{12}$/;
    return aadhaarRegex.test(aadhaar.replace(/\s/g, ''));
  };

  const validatePIN = (pin: string): boolean => {
    // Indian PIN: 6 digits
    const pinRegex = /^[0-9]{6}$/;
    return pinRegex.test(pin);
  };

  const validateAge = (dob: string): boolean => {
    const birthDate = new Date(dob);
    const today = new Date();
    const age = today.getFullYear() - birthDate.getFullYear();
    const monthDiff = today.getMonth() - birthDate.getMonth();

    if (monthDiff < 0 || (monthDiff === 0 && today.getDate() < birthDate.getDate())) {
      return age - 1 >= 18;
    }
    return age >= 18;
  };

  const validateForm = (): boolean => {
    const newErrors: Record<string, string> = {};

    if (!formData.pan || !validatePAN(formData.pan)) {
      newErrors.pan = 'Invalid PAN format (e.g., ABCDE1234F)';
    }

    if (!formData.aadhaar || !validateAadhaar(formData.aadhaar)) {
      newErrors.aadhaar = 'Invalid Aadhaar number (12 digits)';
    }

    if (!formData.date_of_birth) {
      newErrors.date_of_birth = 'Date of birth is required';
    } else if (!validateAge(formData.date_of_birth)) {
      newErrors.date_of_birth = 'You must be at least 18 years old';
    }

    if (!formData.address.street) {
      newErrors.street = 'Street address is required';
    }

    if (!formData.address.city) {
      newErrors.city = 'City is required';
    }

    if (!formData.address.state) {
      newErrors.state = 'State is required';
    }

    if (!formData.address.pin || !validatePIN(formData.address.pin)) {
      newErrors.pin = 'Invalid PIN code (6 digits)';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!validateForm()) {
      return;
    }

    setIsSubmitting(true);
    setError(null);

    try {
      // Submit KYC
      await api.updateKYC({
        ...formData,
        pan: formData.pan.toUpperCase(),
        aadhaar: formData.aadhaar.replace(/\s/g, ''), // Remove spaces
      });

      setSuccess(true);

      // Redirect to dashboard after 2 seconds
      setTimeout(() => {
        navigate('/');
      }, 2000);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to submit KYC. Please try again.');
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleInputChange = (field: string, value: string) => {
    if (field.startsWith('address.')) {
      const addressField = field.split('.')[1];
      setFormData({
        ...formData,
        address: {
          ...formData.address,
          [addressField]: value,
        },
      });
    } else {
      setFormData({
        ...formData,
        [field]: value,
      });
    }

    // Clear error for this field
    if (errors[field]) {
      setErrors({ ...errors, [field]: '' });
    }
  };

  if (success) {
    return (
      <AppLayout title="KYC Submitted">
        <div className="max-w-md mx-auto px-4 py-6 flex items-center justify-center min-h-[60vh]">
          <Card className="w-full text-center py-8">
            <div className="text-6xl mb-4">✅</div>
            <h2 className="text-2xl font-bold text-[var(--text-primary)] mb-2">KYC Submitted Successfully!</h2>
            <p className="text-[var(--text-secondary)] mb-4">
              Your KYC documents are under review. We'll notify you once verified.
            </p>
            <p className="text-sm text-[var(--text-muted)]">Redirecting to dashboard...</p>
          </Card>
        </div>
      </AppLayout>
    );
  }

  return (
    <AppLayout title="Complete KYC" showBack>
      <div className="max-w-2xl mx-auto px-4 py-6 space-y-6">
        {/* Page Description */}
        <p className="text-[var(--text-secondary)]">
          Please provide your details to verify your identity and activate your account.
        </p>

        {/* Error Alert */}
        {error && (
          <Alert variant="error" onDismiss={() => setError(null)}>
            {error}
          </Alert>
        )}

        {/* KYC Form */}
        <Card padding="lg">
          <form onSubmit={handleSubmit} className="space-y-6">
            {/* PAN Card */}
            <FormField
              label="PAN Card Number"
              htmlFor="pan"
              error={errors.pan}
              hint="Format: 5 letters, 4 digits, 1 letter (e.g., ABCDE1234F)"
              required
            >
              <Input
                type="text"
                id="pan"
                value={formData.pan}
                onChange={e => handleInputChange('pan', e.target.value)}
                className="uppercase"
                placeholder="ABCDE1234F"
                maxLength={10}
                disabled={isSubmitting}
                error={!!errors.pan}
              />
            </FormField>

            {/* Aadhaar Number */}
            <FormField
              label="Aadhaar Number"
              htmlFor="aadhaar"
              error={errors.aadhaar}
              hint="12-digit Aadhaar number (spaces optional)"
              required
            >
              <Input
                type="text"
                id="aadhaar"
                value={formData.aadhaar}
                onChange={e => handleInputChange('aadhaar', e.target.value)}
                placeholder="1234 5678 9012"
                maxLength={14}
                disabled={isSubmitting}
                error={!!errors.aadhaar}
              />
            </FormField>

            {/* Date of Birth */}
            <FormField
              label="Date of Birth"
              htmlFor="dob"
              error={errors.date_of_birth}
              hint="You must be at least 18 years old"
              required
            >
              <Input
                type="date"
                id="dob"
                value={formData.date_of_birth}
                onChange={e => handleInputChange('date_of_birth', e.target.value)}
                max={new Date().toISOString().split('T')[0]}
                disabled={isSubmitting}
                error={!!errors.date_of_birth}
              />
            </FormField>

            {/* Address Section */}
            <div className="border-t border-[var(--border-default)] pt-6">
              <h3 className="text-lg font-semibold text-[var(--text-primary)] mb-4">Address Details</h3>

              {/* Street */}
              <FormField
                label="Street Address"
                htmlFor="street"
                error={errors.street}
                className="mb-4"
                required
              >
                <Input
                  type="text"
                  id="street"
                  value={formData.address.street}
                  onChange={e => handleInputChange('address.street', e.target.value)}
                  placeholder="123 Main Street, Apartment 4B"
                  disabled={isSubmitting}
                  error={!!errors.street}
                />
              </FormField>

              {/* City & State */}
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4">
                <FormField
                  label="City"
                  htmlFor="city"
                  error={errors.city}
                  required
                >
                  <Input
                    type="text"
                    id="city"
                    value={formData.address.city}
                    onChange={e => handleInputChange('address.city', e.target.value)}
                    placeholder="Mumbai"
                    disabled={isSubmitting}
                    error={!!errors.city}
                  />
                </FormField>

                <FormField
                  label="State"
                  htmlFor="state"
                  error={errors.state}
                  required
                >
                  <Input
                    type="text"
                    id="state"
                    value={formData.address.state}
                    onChange={e => handleInputChange('address.state', e.target.value)}
                    placeholder="Maharashtra"
                    disabled={isSubmitting}
                    error={!!errors.state}
                  />
                </FormField>
              </div>

              {/* PIN Code */}
              <FormField
                label="PIN Code"
                htmlFor="pin"
                error={errors.pin}
                required
              >
                <Input
                  type="text"
                  id="pin"
                  value={formData.address.pin}
                  onChange={e => handleInputChange('address.pin', e.target.value)}
                  placeholder="400001"
                  maxLength={6}
                  disabled={isSubmitting}
                  error={!!errors.pin}
                />
              </FormField>
            </div>

            {/* Privacy Notice */}
            <Alert variant="info">
              <strong>Privacy Notice:</strong> Your information is encrypted and stored securely.
              We use this data only for identity verification as required by regulations.
            </Alert>

            {/* Submit Button */}
            <Button
              type="submit"
              className="w-full"
              size="lg"
              loading={isSubmitting}
            >
              Submit KYC
            </Button>
          </form>
        </Card>
      </div>
    </AppLayout>
  );
}
