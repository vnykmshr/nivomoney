import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useWalletStore } from '../stores/walletStore';
import { api } from '../lib/api';
import { formatCurrency, toPaise } from '../lib/utils';
import type { UPIDepositResponse } from '../types';

type Step = 'input' | 'payment' | 'success';

export function AddMoney() {
  const navigate = useNavigate();
  const { wallets, fetchWallets } = useWalletStore();

  // Form state
  const [step, setStep] = useState<Step>('input');
  const [walletId, setWalletId] = useState('');
  const [amount, setAmount] = useState('');

  // Payment state
  const [depositResponse, setDepositResponse] = useState<UPIDepositResponse | null>(null);
  const [isInitiating, setIsInitiating] = useState(false);
  const [isCompleting, setIsCompleting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [errors, setErrors] = useState<Record<string, string>>({});

  useEffect(() => {
    if (wallets.length === 0) {
      fetchWallets().catch(err => console.error('Failed to fetch wallets:', err));
    }
  }, [wallets.length, fetchWallets]);

  // Auto-select first active wallet
  useEffect(() => {
    if (!walletId && wallets.length > 0) {
      const activeWallet = wallets.find(w => w.status === 'active');
      if (activeWallet) {
        setWalletId(activeWallet.id);
      }
    }
  }, [wallets, walletId]);

  const selectedWallet = wallets.find(w => w.id === walletId);

  // Validate form
  const validateForm = (): boolean => {
    const newErrors: Record<string, string> = {};

    if (!walletId) {
      newErrors.walletId = 'Please select a wallet';
    }

    if (!amount) {
      newErrors.amount = 'Please enter an amount';
    } else {
      const amountNum = parseFloat(amount);
      if (isNaN(amountNum) || amountNum <= 0) {
        newErrors.amount = 'Amount must be greater than 0';
      } else if (amountNum < 1) {
        newErrors.amount = 'Minimum amount is ₹1';
      } else if (amountNum > 100000) {
        newErrors.amount = 'Maximum amount is ₹100,000';
      }
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleInitiateDeposit = async () => {
    if (!validateForm()) return;

    setIsInitiating(true);
    setError(null);

    try {
      const response = await api.initiateUPIDeposit({
        wallet_id: walletId,
        amount: toPaise(parseFloat(amount)),
        currency: 'INR',
        description: 'Add money via UPI',
      });

      setDepositResponse(response);
      setStep('payment');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to initiate deposit');
    } finally {
      setIsInitiating(false);
    }
  };

  const handleSimulatePayment = async () => {
    if (!depositResponse) return;

    setIsCompleting(true);
    setError(null);

    try {
      // Generate a mock UPI transaction ID
      const upiTxId = `UPI${Date.now()}${Math.random().toString(36).substr(2, 9)}`;

      await api.completeUPIDeposit({
        transaction_id: depositResponse.transaction.id,
        upi_transaction_id: upiTxId,
        status: 'success',
      });

      setStep('success');

      // Refetch wallets to show updated balance
      await fetchWallets();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to complete deposit');
    } finally {
      setIsCompleting(false);
    }
  };

  const handleAddMore = () => {
    setStep('input');
    setAmount('');
    setDepositResponse(null);
    setError(null);
    setErrors({});
  };

  // Input Step
  if (step === 'input') {
    return (
      <div className="min-h-screen bg-gray-50">
        <nav className="bg-white shadow-sm">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            <div className="flex justify-between h-16 items-center">
              <h1 className="text-xl font-bold text-primary-600">Nivo Money</h1>
              <button onClick={() => navigate('/dashboard')} className="btn-secondary">
                Back to Dashboard
              </button>
            </div>
          </div>
        </nav>

        <main className="max-w-2xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <div className="card">
            <h2 className="text-2xl font-bold mb-6">Add Money</h2>

            {error && (
              <div className="mb-4 p-4 bg-red-100 text-red-800 rounded-lg">
                {error}
              </div>
            )}

            <div className="space-y-6">
              {/* Wallet Selection */}
              <div>
                <label htmlFor="wallet" className="block text-sm font-medium text-gray-700 mb-2">
                  To Wallet
                </label>
                <select
                  id="wallet"
                  value={walletId}
                  onChange={e => {
                    setWalletId(e.target.value);
                    setErrors({ ...errors, walletId: '' });
                  }}
                  className="input-field"
                >
                  <option value="">Select a wallet</option>
                  {wallets
                    .filter(w => w.status === 'active')
                    .map(wallet => (
                      <option key={wallet.id} value={wallet.id}>
                        {wallet.currency} Wallet - {formatCurrency(wallet.available_balance)}
                      </option>
                    ))}
                </select>
                {errors.walletId && (
                  <p className="text-sm text-red-600 mt-1">{errors.walletId}</p>
                )}
                {selectedWallet && (
                  <p className="text-sm text-gray-600 mt-1">
                    Current Balance: {formatCurrency(selectedWallet.available_balance)}
                  </p>
                )}
              </div>

              {/* Amount */}
              <div>
                <label htmlFor="amount" className="block text-sm font-medium text-gray-700 mb-2">
                  Amount (₹)
                </label>
                <input
                  type="number"
                  id="amount"
                  value={amount}
                  onChange={e => {
                    setAmount(e.target.value);
                    setErrors({ ...errors, amount: '' });
                  }}
                  className="input-field"
                  placeholder="0.00"
                  step="0.01"
                  min="1"
                  max="100000"
                />
                {errors.amount && (
                  <p className="text-sm text-red-600 mt-1">{errors.amount}</p>
                )}
                <p className="text-xs text-gray-500 mt-1">
                  Min: ₹1 | Max: ₹100,000
                </p>
              </div>

              {/* Quick Amount Buttons */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Quick Select
                </label>
                <div className="grid grid-cols-4 gap-2">
                  {[100, 500, 1000, 5000].map(amt => (
                    <button
                      key={amt}
                      type="button"
                      onClick={() => setAmount(amt.toString())}
                      className="btn-secondary py-2 text-sm"
                    >
                      ₹{amt}
                    </button>
                  ))}
                </div>
              </div>

              {/* Continue Button */}
              <button
                type="button"
                onClick={handleInitiateDeposit}
                disabled={isInitiating || !walletId || !amount}
                className="btn-primary w-full"
              >
                {isInitiating ? 'Initiating...' : 'Continue'}
              </button>
            </div>
          </div>
        </main>
      </div>
    );
  }

  // Payment Step
  if (step === 'payment' && depositResponse) {
    return (
      <div className="min-h-screen bg-gray-50">
        <nav className="bg-white shadow-sm">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            <div className="flex justify-between h-16 items-center">
              <h1 className="text-xl font-bold text-primary-600">Nivo Money</h1>
            </div>
          </div>
        </nav>

        <main className="max-w-2xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <div className="card">
            <h2 className="text-2xl font-bold mb-6">Complete Payment</h2>

            {error && (
              <div className="mb-4 p-4 bg-red-100 text-red-800 rounded-lg">
                {error}
              </div>
            )}

            <div className="space-y-6">
              {/* Amount Display */}
              <div className="text-center bg-primary-50 rounded-lg p-6">
                <p className="text-sm text-gray-600 mb-2">Amount to Pay</p>
                <p className="text-4xl font-bold text-primary-600">
                  {formatCurrency(depositResponse.transaction.amount)}
                </p>
              </div>

              {/* Virtual UPI ID */}
              <div className="bg-gray-50 rounded-lg p-4">
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Pay to this UPI ID
                </label>
                <div className="flex items-center gap-2">
                  <input
                    type="text"
                    value={depositResponse.virtual_upi_id}
                    readOnly
                    className="input-field flex-1 bg-white font-mono text-sm"
                  />
                  <button
                    type="button"
                    onClick={() => {
                      navigator.clipboard.writeText(depositResponse.virtual_upi_id);
                    }}
                    className="btn-secondary whitespace-nowrap"
                  >
                    Copy
                  </button>
                </div>
              </div>

              {/* QR Code (Mock) */}
              <div className="text-center">
                <p className="text-sm text-gray-600 mb-3">Or scan this QR code</p>
                <div className="inline-block p-4 bg-white border-2 border-gray-200 rounded-lg">
                  <img
                    src={depositResponse.qr_code}
                    alt="UPI QR Code"
                    className="w-48 h-48 mx-auto"
                  />
                </div>
              </div>

              {/* Instructions */}
              <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
                <h3 className="font-semibold text-blue-900 mb-2">Payment Instructions</h3>
                <ol className="list-decimal list-inside space-y-1 text-sm text-blue-800">
                  {depositResponse.instructions.map((instruction, idx) => (
                    <li key={idx}>{instruction}</li>
                  ))}
                </ol>
              </div>

              {/* Expiry Warning */}
              <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4">
                <p className="text-sm text-yellow-800">
                  ⏰ This payment link expires at{' '}
                  <span className="font-semibold">
                    {new Date(depositResponse.expires_at).toLocaleTimeString()}
                  </span>
                </p>
              </div>

              {/* Simulate Payment Button (for demo) */}
              <div className="border-t border-gray-200 pt-6">
                <button
                  type="button"
                  onClick={handleSimulatePayment}
                  disabled={isCompleting}
                  className="btn-primary w-full"
                >
                  {isCompleting ? 'Processing...' : '✨ Simulate Payment (Demo)'}
                </button>
                <p className="text-xs text-gray-500 text-center mt-2">
                  In production, this would happen automatically when you complete payment in your UPI app
                </p>
              </div>

              <button
                type="button"
                onClick={() => navigate('/dashboard')}
                className="btn-secondary w-full"
              >
                Cancel
              </button>
            </div>
          </div>
        </main>
      </div>
    );
  }

  // Success Step
  if (step === 'success' && depositResponse) {
    return (
      <div className="min-h-screen bg-gray-50">
        <nav className="bg-white shadow-sm">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            <div className="flex justify-between h-16 items-center">
              <h1 className="text-xl font-bold text-primary-600">Nivo Money</h1>
            </div>
          </div>
        </nav>

        <main className="max-w-2xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <div className="card">
            <div className="text-center">
              {/* Success Icon */}
              <div className="w-16 h-16 bg-green-100 rounded-full flex items-center justify-center mx-auto mb-4">
                <svg className="w-8 h-8 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                </svg>
              </div>

              <h2 className="text-2xl font-bold text-gray-900 mb-2">Money Added Successfully!</h2>
              <p className="text-gray-600 mb-6">
                Your wallet has been credited with{' '}
                <span className="font-semibold text-primary-600">
                  {formatCurrency(depositResponse.transaction.amount)}
                </span>
              </p>

              {/* Transaction Details */}
              <div className="bg-gray-50 rounded-lg p-4 mb-6 text-left">
                <h3 className="font-semibold text-gray-900 mb-3">Transaction Details</h3>
                <div className="space-y-2 text-sm">
                  <div className="flex justify-between">
                    <span className="text-gray-600">Transaction ID:</span>
                    <span className="font-mono text-gray-900">{depositResponse.transaction.id.substring(0, 12)}...</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-gray-600">UPI ID:</span>
                    <span className="font-mono text-gray-900">{depositResponse.virtual_upi_id}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-gray-600">Status:</span>
                    <span className="text-green-600 font-semibold">✓ Completed</span>
                  </div>
                </div>
              </div>

              {/* Action Buttons */}
              <div className="space-y-3">
                <button
                  type="button"
                  onClick={handleAddMore}
                  className="btn-primary w-full"
                >
                  Add More Money
                </button>
                <button
                  type="button"
                  onClick={() => navigate('/dashboard')}
                  className="btn-secondary w-full"
                >
                  Back to Dashboard
                </button>
              </div>
            </div>
          </div>
        </main>
      </div>
    );
  }

  return null;
}
