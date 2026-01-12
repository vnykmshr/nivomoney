import { useState, useEffect } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { useWalletStore } from '../stores/walletStore';
import { api } from '../lib/api';
import { formatCurrency, toPaise, formatDate } from '../lib/utils';
import { AppLayout } from '../components';
import {
  Card,
  Button,
  Input,
  FormField,
  Select,
  Alert,
  Avatar,
  Badge,
  StepIndicator,
  AmountDisplay,
  SuccessState,
  TrustBadge,
} from '../../../shared/components';
import type { Transaction, User } from '@nivo/shared';

type Step = 'input' | 'confirm' | 'receipt';

const STEPS = ['Enter Details', 'Confirm', 'Done'];

export function SendMoney() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const { wallets, fetchWallets } = useWalletStore();

  const [step, setStep] = useState<Step>('input');
  const [sourceWalletId, setSourceWalletId] = useState('');
  const [nivoAddress, setNivoAddress] = useState('');
  const [amount, setAmount] = useState('');
  const [description, setDescription] = useState('');

  const [recipient, setRecipient] = useState<User | null>(null);
  const [recipientWalletId, setRecipientWalletId] = useState('');
  const [lookingUpRecipient, setLookingUpRecipient] = useState(false);
  const [beneficiaryId, setBeneficiaryId] = useState<string | null>(null);

  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [transaction, setTransaction] = useState<Transaction | null>(null);
  const [errors, setErrors] = useState<Record<string, string>>({});

  const currentStepIndex = step === 'input' ? 0 : step === 'confirm' ? 1 : 2;

  useEffect(() => {
    if (wallets.length === 0) {
      fetchWallets().catch(err => console.error('Failed to fetch wallets:', err));
    }
  }, [wallets.length, fetchWallets]);

  useEffect(() => {
    if (!sourceWalletId && wallets.length > 0) {
      const defaultWallet = wallets.find(w => w.type === 'default' && w.status === 'active');
      const activeWallet = wallets.find(w => w.status === 'active');
      const selected = defaultWallet || activeWallet;
      if (selected) setSourceWalletId(selected.id);
    }
  }, [wallets, sourceWalletId]);

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
        const phone = beneficiary.phone.replace('+91', '');
        setNivoAddress(`${phone}@nivomoney`);
        setDescription(`Transfer to ${beneficiary.nickname}`);
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
    } catch {
      setErrors({ nivoAddress: 'Failed to load beneficiary details' });
    } finally {
      setLookingUpRecipient(false);
    }
  };

  const parsePhoneFromAddress = (address: string): string | null => {
    const match = address.match(/^(\d{10})@nivomoney$/i);
    if (!match) return null;
    return `+91${match[1]}`;
  };

  const isValidNivoAddress = (address: string): boolean => {
    return /^\d{10}@nivomoney$/i.test(address);
  };

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
      const user = await api.lookupUser(phone);
      setRecipient(user);

      const walletsResp = await fetch(`/api/v1/wallet/wallets?user_id=${user.id}`, {
        headers: { 'Authorization': `Bearer ${localStorage.getItem('auth_token')}` },
      });

      if (walletsResp.ok) {
        const walletsData = await walletsResp.json();
        const userWallets = walletsData.data || walletsData;
        const activeWallet = userWallets.find((w: { status: string; currency: string }) =>
          w.status === 'active' && w.currency === 'INR'
        );

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

  const validateInput = (): boolean => {
    const newErrors: Record<string, string> = {};

    if (!sourceWalletId) newErrors.sourceWalletId = 'Please select a source wallet';
    if (!nivoAddress) {
      newErrors.nivoAddress = 'Please enter recipient Nivo Money address';
    } else if (!isValidNivoAddress(nivoAddress)) {
      newErrors.nivoAddress = 'Invalid format. Use: 9876543210@nivomoney';
    }
    if (!recipient) newErrors.nivoAddress = 'Please lookup recipient first';

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
    if (validateInput()) setStep('confirm');
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

  const walletOptions = wallets
    .filter(w => w.status === 'active')
    .map(w => ({
      value: w.id,
      label: `${w.currency} Wallet - ${formatCurrency(w.available_balance)}`,
    }));

  // Input Step
  if (step === 'input') {
    return (
      <AppLayout title="Send Money" showBack>
        <div className="max-w-lg mx-auto px-4 py-6">
          {/* Step Indicator */}
          <div className="mb-6">
            <StepIndicator steps={STEPS} currentStep={currentStepIndex} variant="numbered" />
          </div>

          <Card padding="lg">
            {error && <Alert variant="error" className="mb-4">{error}</Alert>}

            <div className="space-y-5">
              <FormField label="From Wallet" htmlFor="sourceWallet" error={errors.sourceWalletId} required>
                <Select
                  id="sourceWallet"
                  value={sourceWalletId}
                  onChange={e => setSourceWalletId(e.target.value)}
                  options={[{ value: '', label: 'Select a wallet' }, ...walletOptions]}
                  error={!!errors.sourceWalletId}
                />
                {selectedWallet && (
                  <p className="text-sm text-[var(--text-muted)] mt-1">
                    Available: {formatCurrency(selectedWallet.available_balance)}
                  </p>
                )}
              </FormField>

              <FormField
                label="To (Nivo Money Address)"
                htmlFor="nivoAddress"
                error={errors.nivoAddress}
                hint="Format: phone@nivomoney (e.g., 9876543210@nivomoney)"
                required
              >
                <div className="flex gap-2">
                  <Input
                    id="nivoAddress"
                    value={nivoAddress}
                    onChange={e => {
                      setNivoAddress(e.target.value.toLowerCase());
                      setRecipient(null);
                      setErrors({ ...errors, nivoAddress: '' });
                    }}
                    placeholder="9876543210@nivomoney"
                    className="flex-1"
                    error={!!errors.nivoAddress}
                  />
                  <Button
                    variant="secondary"
                    onClick={lookupRecipient}
                    disabled={lookingUpRecipient || !nivoAddress}
                    loading={lookingUpRecipient}
                  >
                    Verify
                  </Button>
                </div>
              </FormField>

              {recipient && (
                <div className="p-3 bg-[var(--surface-success)] border border-[var(--color-success-200)] rounded-lg">
                  <div className="flex items-center gap-3">
                    <Avatar name={recipient.full_name} size="sm" />
                    <div>
                      <p className="font-medium text-[var(--text-success)]">{recipient.full_name}</p>
                      <p className="text-sm text-[var(--color-success-600)]">Verified Nivo Money user</p>
                    </div>
                  </div>
                </div>
              )}

              <FormField label="Amount (₹)" htmlFor="amount" error={errors.amount} required>
                <Input
                  id="amount"
                  type="number"
                  value={amount}
                  onChange={e => {
                    setAmount(e.target.value);
                    setErrors({ ...errors, amount: '' });
                  }}
                  placeholder="0.00"
                  error={!!errors.amount}
                />
              </FormField>

              <FormField label="Note (Optional)" htmlFor="description">
                <Input
                  id="description"
                  value={description}
                  onChange={e => setDescription(e.target.value)}
                  placeholder="What's this for?"
                  maxLength={100}
                />
              </FormField>

              <Button
                className="w-full"
                size="lg"
                onClick={handleContinueToConfirm}
                disabled={!recipient || !amount}
              >
                Continue
              </Button>
            </div>
          </Card>

          {/* Trust Badge */}
          <div className="flex justify-center mt-6">
            <TrustBadge variant="encrypted" size="sm" theme="light" />
          </div>
        </div>
      </AppLayout>
    );
  }

  // Confirmation Step
  if (step === 'confirm') {
    const amountPaise = parseFloat(amount) * 100;

    return (
      <AppLayout title="Confirm Transfer" showBack>
        <div className="max-w-lg mx-auto px-4 py-6">
          {/* Step Indicator */}
          <div className="mb-6">
            <StepIndicator steps={STEPS} currentStep={currentStepIndex} variant="numbered" />
          </div>

          <Card padding="lg">
            {error && <Alert variant="error" className="mb-4">{error}</Alert>}

            <div className="space-y-6">
              {/* Amount Display */}
              <AmountDisplay
                amount={amountPaise}
                label="You're sending"
                size="xl"
                variant="highlight"
                showBackground
              />

              {/* Transfer Details */}
              <div className="space-y-3">
                <div className="flex justify-between py-3 border-b border-[var(--border-subtle)]">
                  <span className="text-[var(--text-secondary)]">To</span>
                  <div className="text-right">
                    <p className="font-medium text-[var(--text-primary)]">{recipient?.full_name}</p>
                    <p className="text-sm text-[var(--text-muted)]">{nivoAddress}</p>
                  </div>
                </div>

                <div className="flex justify-between py-3 border-b border-[var(--border-subtle)]">
                  <span className="text-[var(--text-secondary)]">From</span>
                  <span className="font-medium text-[var(--text-primary)]">{selectedWallet?.currency} Wallet</span>
                </div>

                {description && (
                  <div className="flex justify-between py-3 border-b border-[var(--border-subtle)]">
                    <span className="text-[var(--text-secondary)]">Note</span>
                    <span className="font-medium text-[var(--text-primary)]">{description}</span>
                  </div>
                )}
              </div>

              {selectedWallet && (
                <Alert variant="info">
                  Balance after transfer:{' '}
                  <span className="font-semibold">
                    {formatCurrency(selectedWallet.available_balance - amountPaise)}
                  </span>
                </Alert>
              )}

              <div className="flex gap-3">
                <Button variant="secondary" className="flex-1" onClick={() => setStep('input')} disabled={isLoading}>
                  Back
                </Button>
                <Button className="flex-1" onClick={handleConfirmTransfer} loading={isLoading}>
                  Confirm & Send
                </Button>
              </div>
            </div>
          </Card>

          {/* Trust Badge */}
          <div className="flex justify-center mt-6">
            <TrustBadge variant="security" size="sm" theme="light" />
          </div>
        </div>
      </AppLayout>
    );
  }

  // Receipt Step - Success State
  const successDetails = transaction ? [
    { label: 'Transaction ID', value: `${transaction.id.slice(0, 12)}...` },
    { label: 'Amount', value: formatCurrency(transaction.amount) },
    { label: 'To', value: recipient?.full_name || nivoAddress },
    { label: 'Date & Time', value: formatDate(transaction.created_at) },
  ] : [];

  return (
    <AppLayout>
      <div className="max-w-lg mx-auto px-4 py-6">
        {/* Step Indicator */}
        <div className="mb-6">
          <StepIndicator steps={STEPS} currentStep={currentStepIndex} variant="numbered" />
        </div>

        <Card padding="lg">
          <SuccessState
            title="Transfer Successful!"
            message={`${formatCurrency(parseFloat(amount) * 100)} sent to ${recipient?.full_name}`}
            icon="transfer"
            showAnimation
            details={successDetails}
            primaryAction={{
              label: 'Back to Dashboard',
              onClick: () => navigate('/dashboard'),
            }}
            secondaryAction={{
              label: 'Send Again',
              onClick: handleStartNewTransfer,
            }}
          />

          {transaction && (
            <div className="mt-4 flex justify-center">
              <Badge variant="success">{transaction.status.toUpperCase()}</Badge>
            </div>
          )}
        </Card>
      </div>
    </AppLayout>
  );
}
