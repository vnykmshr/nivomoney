import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { api } from '../lib/api';
import type { Beneficiary } from '../types';

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
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto"></div>
          <p className="mt-4 text-gray-600">Loading beneficiaries...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50 py-8">
      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8">
        {/* Header */}
        <div className="flex items-center justify-between mb-6">
          <div>
            <button
              onClick={() => navigate('/dashboard')}
              className="text-blue-600 hover:text-blue-700 mb-2 flex items-center"
            >
              ← Back to Dashboard
            </button>
            <h1 className="text-2xl font-bold text-gray-900">Saved Recipients</h1>
            <p className="text-gray-600 mt-1">Manage your frequently used beneficiaries</p>
          </div>
          <button
            onClick={openAddModal}
            className="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition-colors"
          >
            + Add Beneficiary
          </button>
        </div>

        {/* Error Display */}
        {error && (
          <div className="bg-red-50 border border-red-200 rounded-lg p-4 mb-6">
            <p className="text-red-800">{error}</p>
          </div>
        )}

        {/* Beneficiaries List */}
        {beneficiaries.length === 0 ? (
          <div className="bg-white rounded-lg shadow p-12 text-center">
            <svg
              className="mx-auto h-12 w-12 text-gray-400"
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
            <h3 className="mt-2 text-sm font-medium text-gray-900">No beneficiaries</h3>
            <p className="mt-1 text-sm text-gray-500">
              Add your first beneficiary to make transfers easier
            </p>
            <button
              onClick={openAddModal}
              className="mt-4 bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition-colors"
            >
              Add Beneficiary
            </button>
          </div>
        ) : (
          <div className="bg-white rounded-lg shadow overflow-hidden">
            <ul className="divide-y divide-gray-200">
              {beneficiaries.map((beneficiary) => (
                <li key={beneficiary.id} className="p-4 hover:bg-gray-50 transition-colors">
                  <div className="flex items-center justify-between">
                    <div className="flex-1">
                      <h3 className="text-lg font-medium text-gray-900">{beneficiary.nickname}</h3>
                      <p className="text-sm text-gray-500">{beneficiary.phone}</p>
                    </div>
                    <div className="flex items-center gap-2">
                      <button
                        onClick={() => navigate(`/send-money?beneficiary=${beneficiary.id}`)}
                        className="bg-green-600 text-white px-4 py-2 rounded-lg hover:bg-green-700 transition-colors"
                      >
                        Send Money
                      </button>
                      <button
                        onClick={() => openEditModal(beneficiary)}
                        className="bg-gray-100 text-gray-700 px-4 py-2 rounded-lg hover:bg-gray-200 transition-colors"
                      >
                        Edit
                      </button>
                      <button
                        onClick={() => handleDeleteBeneficiary(beneficiary.id)}
                        className="bg-red-100 text-red-700 px-4 py-2 rounded-lg hover:bg-red-200 transition-colors"
                      >
                        Delete
                      </button>
                    </div>
                  </div>
                </li>
              ))}
            </ul>
          </div>
        )}

        {/* Add Beneficiary Modal */}
        {showAddModal && (
          <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
            <div className="bg-white rounded-lg max-w-md w-full p-6">
              <h2 className="text-xl font-bold text-gray-900 mb-4">Add Beneficiary</h2>
              <form onSubmit={handleAddBeneficiary}>
                <div className="space-y-4">
                  {formErrors.general && (
                    <div className="bg-red-50 border border-red-200 rounded p-3 text-sm text-red-800">
                      {formErrors.general}
                    </div>
                  )}
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      Phone Number
                    </label>
                    <input
                      type="text"
                      value={phone}
                      onChange={(e) => setPhone(e.target.value)}
                      placeholder="9876543210"
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                    />
                    {formErrors.phone && (
                      <p className="mt-1 text-sm text-red-600">{formErrors.phone}</p>
                    )}
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      Nickname
                    </label>
                    <input
                      type="text"
                      value={nickname}
                      onChange={(e) => setNickname(e.target.value)}
                      placeholder="e.g., Mom, John - Rent"
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                    />
                    {formErrors.nickname && (
                      <p className="mt-1 text-sm text-red-600">{formErrors.nickname}</p>
                    )}
                  </div>
                </div>
                <div className="mt-6 flex gap-3">
                  <button
                    type="submit"
                    disabled={isSubmitting}
                    className="flex-1 bg-blue-600 text-white py-2 rounded-lg hover:bg-blue-700 transition-colors disabled:opacity-50"
                  >
                    {isSubmitting ? 'Adding...' : 'Add'}
                  </button>
                  <button
                    type="button"
                    onClick={() => setShowAddModal(false)}
                    className="flex-1 bg-gray-100 text-gray-700 py-2 rounded-lg hover:bg-gray-200 transition-colors"
                  >
                    Cancel
                  </button>
                </div>
              </form>
            </div>
          </div>
        )}

        {/* Edit Beneficiary Modal */}
        {showEditModal && selectedBeneficiary && (
          <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
            <div className="bg-white rounded-lg max-w-md w-full p-6">
              <h2 className="text-xl font-bold text-gray-900 mb-4">Edit Beneficiary</h2>
              <form onSubmit={handleUpdateBeneficiary}>
                <div className="space-y-4">
                  {formErrors.general && (
                    <div className="bg-red-50 border border-red-200 rounded p-3 text-sm text-red-800">
                      {formErrors.general}
                    </div>
                  )}
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      Phone Number
                    </label>
                    <input
                      type="text"
                      value={selectedBeneficiary.phone}
                      disabled
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg bg-gray-50 text-gray-500"
                    />
                    <p className="mt-1 text-xs text-gray-500">Phone number cannot be changed</p>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      Nickname
                    </label>
                    <input
                      type="text"
                      value={nickname}
                      onChange={(e) => setNickname(e.target.value)}
                      placeholder="e.g., Mom, John - Rent"
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                    />
                    {formErrors.nickname && (
                      <p className="mt-1 text-sm text-red-600">{formErrors.nickname}</p>
                    )}
                  </div>
                </div>
                <div className="mt-6 flex gap-3">
                  <button
                    type="submit"
                    disabled={isSubmitting}
                    className="flex-1 bg-blue-600 text-white py-2 rounded-lg hover:bg-blue-700 transition-colors disabled:opacity-50"
                  >
                    {isSubmitting ? 'Updating...' : 'Update'}
                  </button>
                  <button
                    type="button"
                    onClick={() => {
                      setShowEditModal(false);
                      setSelectedBeneficiary(null);
                    }}
                    className="flex-1 bg-gray-100 text-gray-700 py-2 rounded-lg hover:bg-gray-200 transition-colors"
                  >
                    Cancel
                  </button>
                </div>
              </form>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
