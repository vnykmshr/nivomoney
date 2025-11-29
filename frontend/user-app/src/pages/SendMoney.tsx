import { useState, useEffect } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { useWalletStore } from '../stores/walletStore';
import { api } from '../lib/api';
import { formatCurrency, toPaise, formatDate } from '../lib/utils';
import type { Transaction, User } from '../types';

type Step = 'input' | 'confirm' | 'receipt';

export function SendMoney() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const { wallets, fetchWallets } = useWalletStore();

  // Form state
  const [step, setStep] = useState<Step>('input');
  const [sourceWalletId, setSourceWalletId] = useState('');
  const [nivoAddress, setNivoAddress] = useState(''); // e.g., 9876543210@nivomoney
  const [amount, setAmount] = useState('');
  const [description, setDescription] = useState('');

  // Recipient state
  const [recipient, setRecipient] = useState<User | null>(null);
  const [recipientWalletId, setRecipientWalletId] = useState('');
  const [lookingUpRecipient, setLookingUpRecipient] = useState(false);
  const [beneficiaryId, setBeneficiaryId] = useState<string | null>(null);

  // Transaction state
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [transaction, setTransaction] = useState<Transaction | null>(null);
  const [errors, setErrors] = useState<Record<string, string>>({});

  useEffect(() => {
    if (wallets.length === 0) {
      fetchWallets().catch(err => console.error('Failed to fetch wallets:', err));
    }
  }, [wallets.length, fetchWallets]);

  // Auto-select default active wallet (prioritize 'default' type)
  useEffect(() => {
    if (!sourceWalletId && wallets.length > 0) {
      const defaultWallet = wallets.find(w => w.type === 'default' && w.status === 'active');
      const activeWallet = wallets.find(w => w.status === 'active');
      const selected = defaultWallet || activeWallet;
      if (selected) {
        setSourceWalletId(selected.id);
      }
    }
  }, [wallets, sourceWalletId]);

  // Pre-fill from beneficiary if provided
  useEffect(() => {
    const beneficiaryIdParam = searchParams.get('beneficiary');
    if (beneficiaryIdParam && !beneficiaryId) {
      setBeneficiaryId(beneficiaryIdParam);
      loadBeneficiary(beneficiaryIdParam);
    }
  }, [searchParams, beneficiaryId]);

  const loadBeneficiary = async (id: string) => {
    try {
      const beneficiaries = await api.getBeneficiaries();
      const beneficiary = beneficiaries.find(b => b.id === id);

      if (beneficiary) {
        // Convert phone to nivo address format
        const phone = beneficiary.phone.replace('+91', '');
        setNivoAddress(`${phone}@nivomoney`);

        // Pre-fill description
        setDescription(`Transfer to ${beneficiary.nickname}`);

        // Lookup the recipient automatically
        lookupRecipientByPhone(beneficiary.phone, beneficiary.wallet_id);
      }
    } catch (err) {
      console.error('Failed to load beneficiary:', err);
    }
  };

  const lookupRecipientByPhone = async (phone: string, walletId: string) => {
    setLookingUpRecipient(true);
    try {
      const user = await api.lookupUser(phone);
      setRecipient(user);
      setRecipientWalletId(walletId);
    } catch (err: any) {
      setErrors({ nivoAddress: 'Failed to load beneficiary details' });
    } finally {
      setLookingUpRecipient(false);
    }
  };

  // Parse phone from nivo address (9876543210@nivomoney -> +919876543210)
  const parsePhoneFromAddress = (address: string): string | null => {
    const match = address.match(/^(\d{10})@nivomoney$/i);
    if (!match) return null;
    return `+91${match[1]}`; // Indian phone format
  };

  // Validate nivo address format
  const isValidNivoAddress = (address: string): boolean => {
    return /^\d{10}@nivomoney$/i.test(address);
  };

  // Lookup recipient by nivo address
  const lookupRecipient = async () => {
    if (!nivoAddress) return;

    if (!isValidNivoAddress(nivoAddress)) {
      setErrors({ ...errors, nivoAddress: 'Invalid format. Use: 9876543210@nivomoney' });
      return;
    }

    const phone = parsePhoneFromAddress(nivoAddress);
    if (!phone) {
      setErrors({ ...errors, nivoAddress: 'Invalid phone number' });
      return;
    }

    setLookingUpRecipient(true);
    setError(null);

    try {
      // Look up user by phone
      const user = await api.lookupUser(phone);
      setRecipient(user);

      // Fetch recipient's wallets to get their active INR wallet
      // Note: We use direct fetch here because we need to query by user_id
      const walletsResp = await fetch(`/api/v1/wallet/wallets?user_id=${user.id}`, {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('auth_token')}`,
        },
      });

      if (walletsResp.ok) {
        const walletsData = await walletsResp.json();
        const userWallets = walletsData.data || walletsData;
        const activeWallet = userWallets.find((w: any) => w.status === 'active' && w.currency === 'INR');

        if (activeWallet) {
          setRecipientWalletId(activeWallet.id);
        } else {
          setErrors({ ...errors, nivoAddress: 'Recipient does not have an active INR wallet' });
          setRecipient(null);
        }
      } else {
        throw new Error('Failed to fetch recipient wallets');
      }
    } catch (err) {
      if (err instanceof Error && err.message.includes('404')) {
        setErrors({ ...errors, nivoAddress: 'No Nivo Money user found with this phone number' });
      } else {
        setError(err instanceof Error ? err.message : 'Failed to lookup recipient');
      }
    } finally {
      setLookingUpRecipient(false);
    }
  };

  // Validate form on input step
  const validateInput = (): boolean => {
    const newErrors: Record<string, string> = {};

    if (!sourceWalletId) {
      newErrors.sourceWalletId = 'Please select a source wallet';
    }

    if (!nivoAddress) {
      newErrors.nivoAddress = 'Please enter recipient Nivo Money address';
    } else if (!isValidNivoAddress(nivoAddress)) {
      newErrors.nivoAddress = 'Invalid format. Use: 9876543210@nivomoney';
    }

    if (!recipient) {
      newErrors.nivoAddress = 'Please lookup recipient first';
    }

    if (!amount) {
      newErrors.amount = 'Please enter an amount';
    } else {
      const amountNum = parseFloat(amount);
      if (isNaN(amountNum) || amountNum <= 0) {
        newErrors.amount = 'Amount must be greater than 0';
      } else if (amountNum < 1) {
        newErrors.amount = 'Minimum amount is ₹1';
      } else {
        const sourceWallet = wallets.find(w => w.id === sourceWalletId);
        if (sourceWallet && amountNum * 100 > sourceWallet.available_balance) {
          newErrors.amount = `Insufficient balance. Available: ${formatCurrency(sourceWallet.available_balance)}`;
        }
      }
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleContinueToConfirm = () => {
    if (validateInput()) {
      setStep('confirm');
    }
  };

  const handleConfirmTransfer = async () => {
    setIsLoading(true);
    setError(null);

    try {
      const tx = await api.createTransfer({
        source_wallet_id: sourceWalletId,
        destination_wallet_id: recipientWalletId,
        amount_paise: toPaise(parseFloat(amount)),
        currency: 'INR',
        description: description || `Sent to ${nivoAddress}`,
      });

      setTransaction(tx);
      setStep('receipt');

      // Refetch wallets to update balance
      await fetchWallets();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to send money');
      setIsLoading(false);
    }
  };

  const handleStartNewTransfer = () => {
    setStep('input');
    setNivoAddress('');
    setAmount('');
    setDescription('');
    setRecipient(null);
    setRecipientWalletId('');
    setTransaction(null);
    setError(null);
    setErrors({});
  };

  const selectedWallet = wallets.find(w => w.id === sourceWalletId);

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
            <h2 className="text-2xl font-bold mb-6">Send Money</h2>

            {error && (
              <div className="mb-4 p-4 bg-red-100 text-red-800 rounded-lg">
                {error}
              </div>
            )}

            <div className="space-y-6">
              {/* Source Wallet */}
              <div>
                <label htmlFor="sourceWallet" className="block text-sm font-medium text-gray-700 mb-2">
                  From Wallet
                </label>
                <select
                  id="sourceWallet"
                  value={sourceWalletId}
                  onChange={e => setSourceWalletId(e.target.value)}
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
                {errors.sourceWalletId && (
                  <p className="text-sm text-red-600 mt-1">{errors.sourceWalletId}</p>
                )}
                {selectedWallet && (
                  <p className="text-sm text-gray-600 mt-1">
                    Available: {formatCurrency(selectedWallet.available_balance)}
                  </p>
                )}
              </div>

              {/* Recipient Nivo Address */}
              <div>
                <label htmlFor="nivoAddress" className="block text-sm font-medium text-gray-700 mb-2">
                  To (Nivo Money Address)
                </label>
                <div className="flex gap-2">
                  <input
                    type="text"
                    id="nivoAddress"
                    value={nivoAddress}
                    onChange={e => {
                      setNivoAddress(e.target.value.toLowerCase());
                      setRecipient(null);
                      setErrors({ ...errors, nivoAddress: '' });
                    }}
                    className="input-field flex-1"
                    placeholder="9876543210@nivomoney"
                  />
                  <button
                    type="button"
                    onClick={lookupRecipient}
                    disabled={lookingUpRecipient || !nivoAddress}
                    className="btn-secondary whitespace-nowrap"
                  >
                    {lookingUpRecipient ? 'Looking up...' : 'Verify'}
                  </button>
                </div>
                {errors.nivoAddress && (
                  <p className="text-sm text-red-600 mt-1">{errors.nivoAddress}</p>
                )}
                <p className="text-xs text-gray-500 mt-1">
                  Format: phone@nivomoney (e.g., 9876543210@nivomoney)
                </p>

                {/* Recipient Card */}
                {recipient && (
                  <div className="mt-3 p-3 bg-green-50 border border-green-200 rounded-lg">
                    <div className="flex items-center gap-2">
                      <div className="w-10 h-10 rounded-full bg-green-600 flex items-center justify-center text-white font-bold">
                        {recipient.full_name.charAt(0).toUpperCase()}
                      </div>
                      <div>
                        <p className="font-medium text-green-900">{recipient.full_name}</p>
                        <p className="text-sm text-green-700">✓ Verified Nivo Money user</p>
                      </div>
                    </div>
                  </div>
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
                />
                {errors.amount && (
                  <p className="text-sm text-red-600 mt-1">{errors.amount}</p>
                )}
              </div>

              {/* Description */}
              <div>
                <label htmlFor="description" className="block text-sm font-medium text-gray-700 mb-2">
                  Note (Optional)
                </label>
                <input
                  type="text"
                  id="description"
                  value={description}
                  onChange={e => setDescription(e.target.value)}
                  className="input-field"
                  placeholder="What's this for?"
                  maxLength={100}
                />
              </div>

              {/* Continue Button */}
              <button
                type="button"
                onClick={handleContinueToConfirm}
                className="btn-primary w-full"
                disabled={!recipient || !amount}
              >
                Continue
              </button>
            </div>
          </div>
        </main>
      </div>
    );
  }

  // Confirmation Step
  if (step === 'confirm') {
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
            <h2 className="text-2xl font-bold mb-6">Confirm Transfer</h2>

            {error && (
              <div className="mb-4 p-4 bg-red-100 text-red-800 rounded-lg">
                {error}
              </div>
            )}

            <div className="space-y-6">
              {/* Transfer Details */}
              <div className="bg-gray-50 rounded-lg p-6 space-y-4">
                <div className="text-center">
                  <p className="text-sm text-gray-600">You're sending</p>
                  <p className="text-4xl font-bold text-gray-900 my-2">
                    {formatCurrency(parseFloat(amount) * 100)}
                  </p>
                </div>

                <div className="border-t border-gray-200 pt-4 space-y-3">
                  <div className="flex justify-between">
                    <span className="text-sm text-gray-600">To</span>
                    <div className="text-right">
                      <p className="font-medium text-gray-900">{recipient?.full_name}</p>
                      <p className="text-sm text-gray-600">{nivoAddress}</p>
                    </div>
                  </div>

                  <div className="flex justify-between">
                    <span className="text-sm text-gray-600">From</span>
                    <span className="font-medium text-gray-900">
                      {selectedWallet?.currency} Wallet
                    </span>
                  </div>

                  {description && (
                    <div className="flex justify-between">
                      <span className="text-sm text-gray-600">Note</span>
                      <span className="font-medium text-gray-900">{description}</span>
                    </div>
                  )}
                </div>
              </div>

              {/* Balance After Transfer */}
              {selectedWallet && (
                <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
                  <p className="text-sm text-blue-900">
                    Balance after transfer:{' '}
                    <span className="font-semibold">
                      {formatCurrency(selectedWallet.available_balance - parseFloat(amount) * 100)}
                    </span>
                  </p>
                </div>
              )}

              {/* Action Buttons */}
              <div className="flex gap-4">
                <button
                  type="button"
                  onClick={() => setStep('input')}
                  className="flex-1 btn-secondary"
                  disabled={isLoading}
                >
                  Back
                </button>
                <button
                  type="button"
                  onClick={handleConfirmTransfer}
                  className="flex-1 btn-primary"
                  disabled={isLoading}
                >
                  {isLoading ? 'Processing...' : 'Confirm & Send'}
                </button>
              </div>
            </div>
          </div>
        </main>
      </div>
    );
  }

  // Receipt Step
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
        <div className="card text-center">
          {/* Success Icon */}
          <div className="w-20 h-20 bg-green-100 rounded-full flex items-center justify-center mx-auto mb-4">
            <svg className="w-12 h-12 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
            </svg>
          </div>

          <h2 className="text-2xl font-bold text-gray-900 mb-2">Transfer Successful!</h2>
          <p className="text-gray-600 mb-6">
            {formatCurrency(parseFloat(amount) * 100)} sent to {recipient?.full_name}
          </p>

          {/* Transaction Details */}
          {transaction && (
            <div className="bg-gray-50 rounded-lg p-6 text-left space-y-3 mb-6">
              <h3 className="font-semibold text-gray-900 mb-3">Transaction Details</h3>

              <div className="flex justify-between text-sm">
                <span className="text-gray-600">Transaction ID</span>
                <span className="font-mono text-gray-900">{transaction.id.slice(0, 12)}...</span>
              </div>

              <div className="flex justify-between text-sm">
                <span className="text-gray-600">Amount</span>
                <span className="font-semibold text-gray-900">{formatCurrency(transaction.amount)}</span>
              </div>

              <div className="flex justify-between text-sm">
                <span className="text-gray-600">To</span>
                <div className="text-right">
                  <p className="font-medium text-gray-900">{recipient?.full_name}</p>
                  <p className="text-xs text-gray-600">{nivoAddress}</p>
                </div>
              </div>

              <div className="flex justify-between text-sm">
                <span className="text-gray-600">Date & Time</span>
                <span className="text-gray-900">{formatDate(transaction.created_at)}</span>
              </div>

              <div className="flex justify-between text-sm">
                <span className="text-gray-600">Status</span>
                <span className="px-2 py-1 text-xs font-semibold rounded bg-green-100 text-green-800">
                  {transaction.status.toUpperCase()}
                </span>
              </div>

              {transaction.description && (
                <div className="flex justify-between text-sm">
                  <span className="text-gray-600">Note</span>
                  <span className="text-gray-900">{transaction.description}</span>
                </div>
              )}
            </div>
          )}

          {/* Action Buttons */}
          <div className="flex gap-4">
            <button
              onClick={handleStartNewTransfer}
              className="flex-1 btn-secondary"
            >
              Send Again
            </button>
            <button
              onClick={() => navigate('/dashboard')}
              className="flex-1 btn-primary"
            >
              Back to Dashboard
            </button>
          </div>
        </div>
      </main>
    </div>
  );
}
