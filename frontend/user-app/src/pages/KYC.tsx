import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { api } from '../lib/api';
import type { UpdateKYCRequest } from '../types';

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
      <div className="min-h-screen bg-gray-50 flex items-center justify-center p-4">
        <div className="card max-w-md w-full text-center py-8">
          <div className="text-6xl mb-4">✅</div>
          <h2 className="text-2xl font-bold text-gray-900 mb-2">KYC Submitted Successfully!</h2>
          <p className="text-gray-600 mb-4">
            Your KYC documents are under review. We'll notify you once verified.
          </p>
          <p className="text-sm text-gray-500">Redirecting to dashboard...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50 py-8">
      <div className="max-w-3xl mx-auto px-4 sm:px-6 lg:px-8">
        {/* Header */}
        <div className="mb-8">
          <button
            onClick={() => navigate('/')}
            className="text-primary-600 hover:text-primary-700 mb-4 flex items-center"
          >
            ← Back to Dashboard
          </button>
          <h1 className="text-3xl font-bold text-gray-900">Complete KYC Verification</h1>
          <p className="text-gray-600 mt-2">
            Please provide your details to verify your identity and activate your account.
          </p>
        </div>

        {/* Error Alert */}
        {error && (
          <div className="mb-6 p-4 bg-red-100 text-red-800 rounded-lg">
            {error}
          </div>
        )}

        {/* KYC Form */}
        <form onSubmit={handleSubmit} className="card space-y-6">
          {/* PAN Card */}
          <div>
            <label htmlFor="pan" className="block text-sm font-medium text-gray-700 mb-2">
              PAN Card Number <span className="text-red-500">*</span>
            </label>
            <input
              type="text"
              id="pan"
              value={formData.pan}
              onChange={e => handleInputChange('pan', e.target.value)}
              className={`input-field uppercase ${errors.pan ? 'border-red-500' : ''}`}
              placeholder="ABCDE1234F"
              maxLength={10}
              disabled={isSubmitting}
            />
            {errors.pan && <p className="mt-1 text-sm text-red-600">{errors.pan}</p>}
            <p className="mt-1 text-sm text-gray-500">
              Format: 5 letters, 4 digits, 1 letter (e.g., ABCDE1234F)
            </p>
          </div>

          {/* Aadhaar Number */}
          <div>
            <label htmlFor="aadhaar" className="block text-sm font-medium text-gray-700 mb-2">
              Aadhaar Number <span className="text-red-500">*</span>
            </label>
            <input
              type="text"
              id="aadhaar"
              value={formData.aadhaar}
              onChange={e => handleInputChange('aadhaar', e.target.value)}
              className={`input-field ${errors.aadhaar ? 'border-red-500' : ''}`}
              placeholder="1234 5678 9012"
              maxLength={14}
              disabled={isSubmitting}
            />
            {errors.aadhaar && <p className="mt-1 text-sm text-red-600">{errors.aadhaar}</p>}
            <p className="mt-1 text-sm text-gray-500">
              12-digit Aadhaar number (spaces optional)
            </p>
          </div>

          {/* Date of Birth */}
          <div>
            <label htmlFor="dob" className="block text-sm font-medium text-gray-700 mb-2">
              Date of Birth <span className="text-red-500">*</span>
            </label>
            <input
              type="date"
              id="dob"
              value={formData.date_of_birth}
              onChange={e => handleInputChange('date_of_birth', e.target.value)}
              className={`input-field ${errors.date_of_birth ? 'border-red-500' : ''}`}
              max={new Date().toISOString().split('T')[0]}
              disabled={isSubmitting}
            />
            {errors.date_of_birth && <p className="mt-1 text-sm text-red-600">{errors.date_of_birth}</p>}
            <p className="mt-1 text-sm text-gray-500">
              You must be at least 18 years old
            </p>
          </div>

          {/* Address Section */}
          <div className="border-t pt-6">
            <h3 className="text-lg font-semibold text-gray-900 mb-4">Address Details</h3>

            {/* Street */}
            <div className="mb-4">
              <label htmlFor="street" className="block text-sm font-medium text-gray-700 mb-2">
                Street Address <span className="text-red-500">*</span>
              </label>
              <input
                type="text"
                id="street"
                value={formData.address.street}
                onChange={e => handleInputChange('address.street', e.target.value)}
                className={`input-field ${errors.street ? 'border-red-500' : ''}`}
                placeholder="123 Main Street, Apartment 4B"
                disabled={isSubmitting}
              />
              {errors.street && <p className="mt-1 text-sm text-red-600">{errors.street}</p>}
            </div>

            {/* City & State */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4">
              <div>
                <label htmlFor="city" className="block text-sm font-medium text-gray-700 mb-2">
                  City <span className="text-red-500">*</span>
                </label>
                <input
                  type="text"
                  id="city"
                  value={formData.address.city}
                  onChange={e => handleInputChange('address.city', e.target.value)}
                  className={`input-field ${errors.city ? 'border-red-500' : ''}`}
                  placeholder="Mumbai"
                  disabled={isSubmitting}
                />
                {errors.city && <p className="mt-1 text-sm text-red-600">{errors.city}</p>}
              </div>

              <div>
                <label htmlFor="state" className="block text-sm font-medium text-gray-700 mb-2">
                  State <span className="text-red-500">*</span>
                </label>
                <input
                  type="text"
                  id="state"
                  value={formData.address.state}
                  onChange={e => handleInputChange('address.state', e.target.value)}
                  className={`input-field ${errors.state ? 'border-red-500' : ''}`}
                  placeholder="Maharashtra"
                  disabled={isSubmitting}
                />
                {errors.state && <p className="mt-1 text-sm text-red-600">{errors.state}</p>}
              </div>
            </div>

            {/* PIN Code */}
            <div>
              <label htmlFor="pin" className="block text-sm font-medium text-gray-700 mb-2">
                PIN Code <span className="text-red-500">*</span>
              </label>
              <input
                type="text"
                id="pin"
                value={formData.address.pin}
                onChange={e => handleInputChange('address.pin', e.target.value)}
                className={`input-field ${errors.pin ? 'border-red-500' : ''}`}
                placeholder="400001"
                maxLength={6}
                disabled={isSubmitting}
              />
              {errors.pin && <p className="mt-1 text-sm text-red-600">{errors.pin}</p>}
            </div>
          </div>

          {/* Privacy Notice */}
          <div className="bg-blue-50 p-4 rounded-lg">
            <p className="text-sm text-blue-800">
              <strong>Privacy Notice:</strong> Your information is encrypted and stored securely.
              We use this data only for identity verification as required by regulations.
            </p>
          </div>

          {/* Submit Button */}
          <button
            type="submit"
            className="btn-primary w-full py-3 text-lg"
            disabled={isSubmitting}
          >
            {isSubmitting ? 'Submitting...' : 'Submit KYC'}
          </button>
        </form>
      </div>
    </div>
  );
}
