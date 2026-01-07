import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { api } from '../lib/api';
import type { Beneficiary } from '../types';
import { AppLayout } from '../components';
import {
  Alert,
  Button,
  Card,
  FormField,
  Input,
} from '../../../shared/components';

export function Beneficiaries() {
  const navigate = useNavigate();
  const [beneficiaries, setBeneficiaries] = useState<Beneficiary[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showAddModal, setShowAddModal] = useState(false);
  const [showEditModal, setShowEditModal] = useState(false);
  const [selectedBeneficiary, setSelectedBeneficiary] = useState<Beneficiary | null>(null);

  // Form state
  const [phone, setPhone] = useState('');
  const [nickname, setNickname] = useState('');
  const [formErrors, setFormErrors] = useState<Record<string, string>>({});
  const [isSubmitting, setIsSubmitting] = useState(false);

  useEffect(() => {
    fetchBeneficiaries();
  }, []);

  const fetchBeneficiaries = async () => {
    try {
      setIsLoading(true);
      const data = await api.getBeneficiaries();
      setBeneficiaries(data);
      setError(null);
    } catch (err: any) {
      setError(err.response?.data?.message || 'Failed to load beneficiaries');
    } finally {
      setIsLoading(false);
    }
  };

  const handleAddBeneficiary = async (e: React.FormEvent) => {
    e.preventDefault();
    setFormErrors({});

    if (!phone) {
      setFormErrors({ phone: 'Phone number is required' });
      return;
    }

    if (!nickname) {
      setFormErrors({ nickname: 'Nickname is required' });
      return;
    }

    // Validate phone format (10 digits or +91 followed by 10 digits)
    const phoneRegex = /^(\+91)?[0-9]{10}$/;
    if (!phoneRegex.test(phone)) {
      setFormErrors({ phone: 'Invalid phone format. Use: 9876543210 or +919876543210' });
      return;
    }

    // Normalize phone to +91 format
    const normalizedPhone = phone.startsWith('+91') ? phone : `+91${phone}`;

    setIsSubmitting(true);
    try {
      await api.addBeneficiary({ phone: normalizedPhone, nickname });
      await fetchBeneficiaries();
      setShowAddModal(false);
      setPhone('');
      setNickname('');
    } catch (err: any) {
      setFormErrors({
        general: err.response?.data?.message || 'Failed to add beneficiary'
      });
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleUpdateBeneficiary = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedBeneficiary) return;

    setFormErrors({});

    if (!nickname) {
      setFormErrors({ nickname: 'Nickname is required' });
      return;
    }

    setIsSubmitting(true);
    try {
      await api.updateBeneficiary(selectedBeneficiary.id, { nickname });
      await fetchBeneficiaries();
      setShowEditModal(false);
      setSelectedBeneficiary(null);
      setNickname('');
    } catch (err: any) {
      setFormErrors({
        general: err.response?.data?.message || 'Failed to update beneficiary'
      });
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleDeleteBeneficiary = async (id: string) => {
    if (!confirm('Are you sure you want to remove this beneficiary?')) return;

    try {
      await api.deleteBeneficiary(id);
      await fetchBeneficiaries();
    } catch (err: any) {
      setError(err.response?.data?.message || 'Failed to delete beneficiary');
    }
  };

  const openEditModal = (beneficiary: Beneficiary) => {
    setSelectedBeneficiary(beneficiary);
    setNickname(beneficiary.nickname);
    setFormErrors({});
    setShowEditModal(true);
  };

  const openAddModal = () => {
    setPhone('');
    setNickname('');
    setFormErrors({});
    setShowAddModal(true);
  };

  if (isLoading) {
    return (
      <AppLayout title="Recipients" showBack>
        <div className="max-w-4xl mx-auto px-4 py-6 flex items-center justify-center min-h-[60vh]">
          <div className="text-center">
            <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-[var(--interactive-primary)] mx-auto"></div>
            <p className="mt-4 text-[var(--text-secondary)]">Loading beneficiaries...</p>
          </div>
        </div>
      </AppLayout>
    );
  }

  return (
    <AppLayout title="Recipients" showBack actions={<Button onClick={openAddModal}>+ Add Beneficiary</Button>}>
      <div className="max-w-4xl mx-auto px-4 py-6 space-y-6">
        {/* Page Description */}
        <p className="text-[var(--text-secondary)]">Manage your frequently used beneficiaries</p>

        {/* Error Display */}
        {error && (
          <Alert variant="error" onDismiss={() => setError(null)}>
            {error}
          </Alert>
        )}

        {/* Beneficiaries List */}
        {beneficiaries.length === 0 ? (
          <Card padding="lg" className="text-center py-12">
            <svg
              className="mx-auto h-12 w-12 text-[var(--text-muted)]"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"
              />
            </svg>
            <h3 className="mt-2 text-sm font-medium text-[var(--text-primary)]">No beneficiaries</h3>
            <p className="mt-1 text-sm text-[var(--text-muted)]">
              Add your first beneficiary to make transfers easier
            </p>
            <Button onClick={openAddModal} className="mt-4">
              Add Beneficiary
            </Button>
          </Card>
        ) : (
          <Card className="overflow-hidden">
            <ul className="divide-y divide-[var(--border-default)]">
              {beneficiaries.map((beneficiary) => (
                <li key={beneficiary.id} className="p-4 hover:bg-[var(--surface-muted)] transition-colors">
                  <div className="flex items-center justify-between">
                    <div className="flex-1">
                      <h3 className="text-lg font-medium text-[var(--text-primary)]">{beneficiary.nickname}</h3>
                      <p className="text-sm text-[var(--text-muted)]">{beneficiary.phone}</p>
                    </div>
                    <div className="flex items-center gap-2">
                      <Button
                        onClick={() => navigate(`/send-money?beneficiary=${beneficiary.id}`)}
                      >
                        Send Money
                      </Button>
                      <Button
                        onClick={() => openEditModal(beneficiary)}
                        variant="secondary"
                      >
                        Edit
                      </Button>
                      <Button
                        onClick={() => handleDeleteBeneficiary(beneficiary.id)}
                        variant="danger"
                      >
                        Delete
                      </Button>
                    </div>
                  </div>
                </li>
              ))}
            </ul>
          </Card>
        )}

        {/* Add Beneficiary Modal */}
        {showAddModal && (
          <div className="fixed inset-0 bg-black/50 flex items-center justify-center p-4 z-50">
            <Card padding="lg" className="max-w-md w-full">
              <h2 className="text-xl font-bold text-[var(--text-primary)] mb-4">Add Beneficiary</h2>
              <form onSubmit={handleAddBeneficiary}>
                <div className="space-y-4">
                  {formErrors.general && (
                    <Alert variant="error">
                      {formErrors.general}
                    </Alert>
                  )}
                  <FormField
                    label="Phone Number"
                    htmlFor="add-phone"
                    error={formErrors.phone}
                    required
                  >
                    <Input
                      type="text"
                      id="add-phone"
                      value={phone}
                      onChange={(e) => setPhone(e.target.value)}
                      placeholder="9876543210"
                      error={!!formErrors.phone}
                    />
                  </FormField>
                  <FormField
                    label="Nickname"
                    htmlFor="add-nickname"
                    error={formErrors.nickname}
                    required
                  >
                    <Input
                      type="text"
                      id="add-nickname"
                      value={nickname}
                      onChange={(e) => setNickname(e.target.value)}
                      placeholder="e.g., Mom, John - Rent"
                      error={!!formErrors.nickname}
                    />
                  </FormField>
                </div>
                <div className="mt-6 flex gap-3">
                  <Button
                    type="submit"
                    disabled={isSubmitting}
                    loading={isSubmitting}
                    className="flex-1"
                  >
                    Add
                  </Button>
                  <Button
                    type="button"
                    variant="secondary"
                    onClick={() => setShowAddModal(false)}
                    className="flex-1"
                  >
                    Cancel
                  </Button>
                </div>
              </form>
            </Card>
          </div>
        )}

        {/* Edit Beneficiary Modal */}
        {showEditModal && selectedBeneficiary && (
          <div className="fixed inset-0 bg-black/50 flex items-center justify-center p-4 z-50">
            <Card padding="lg" className="max-w-md w-full">
              <h2 className="text-xl font-bold text-[var(--text-primary)] mb-4">Edit Beneficiary</h2>
              <form onSubmit={handleUpdateBeneficiary}>
                <div className="space-y-4">
                  {formErrors.general && (
                    <Alert variant="error">
                      {formErrors.general}
                    </Alert>
                  )}
                  <FormField
                    label="Phone Number"
                    htmlFor="edit-phone"
                    hint="Phone number cannot be changed"
                  >
                    <Input
                      type="text"
                      id="edit-phone"
                      value={selectedBeneficiary.phone}
                      disabled
                    />
                  </FormField>
                  <FormField
                    label="Nickname"
                    htmlFor="edit-nickname"
                    error={formErrors.nickname}
                    required
                  >
                    <Input
                      type="text"
                      id="edit-nickname"
                      value={nickname}
                      onChange={(e) => setNickname(e.target.value)}
                      placeholder="e.g., Mom, John - Rent"
                      error={!!formErrors.nickname}
                    />
                  </FormField>
                </div>
                <div className="mt-6 flex gap-3">
                  <Button
                    type="submit"
                    disabled={isSubmitting}
                    loading={isSubmitting}
                    className="flex-1"
                  >
                    Update
                  </Button>
                  <Button
                    type="button"
                    variant="secondary"
                    onClick={() => {
                      setShowEditModal(false);
                      setSelectedBeneficiary(null);
                    }}
                    className="flex-1"
                  >
                    Cancel
                  </Button>
                </div>
              </form>
            </Card>
          </div>
        )}
      </div>
    </AppLayout>
  );
}
