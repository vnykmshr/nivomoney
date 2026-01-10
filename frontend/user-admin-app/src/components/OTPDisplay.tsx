/**
 * OTP Display Component
 * Prominently displays OTP code with countdown timer
 */

import { useState, useEffect } from 'react';
import { Card, Badge } from '../../../shared/components';
import type { Verification } from '../lib/api';

interface OTPDisplayProps {
  verification: Verification;
}

export function OTPDisplay({ verification }: OTPDisplayProps) {
  const [timeRemaining, setTimeRemaining] = useState<number>(0);
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    const calculateRemaining = () => {
      const expiresAt = new Date(verification.expires_at).getTime();
      const now = Date.now();
      const remaining = Math.max(0, Math.floor((expiresAt - now) / 1000));
      setTimeRemaining(remaining);
    };

    calculateRemaining();
    const interval = setInterval(calculateRemaining, 1000);

    return () => clearInterval(interval);
  }, [verification.expires_at]);

  const formatTime = (seconds: number): string => {
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins}:${secs.toString().padStart(2, '0')}`;
  };

  const handleCopy = async () => {
    if (verification.otp_code) {
      try {
        await navigator.clipboard.writeText(verification.otp_code);
        setCopied(true);
        setTimeout(() => setCopied(false), 2000);
      } catch (err) {
        console.error('Failed to copy:', err);
      }
    }
  };

  const isExpired = timeRemaining === 0;
  const isUrgent = timeRemaining > 0 && timeRemaining <= 60;

  const getOperationLabel = (type: string): string => {
    const labels: Record<string, string> = {
      withdraw: 'Withdrawal',
      transfer: 'Money Transfer',
      add_beneficiary: 'Add Beneficiary',
      password_change: 'Password Change',
      profile_update: 'Profile Update',
    };
    return labels[type] || type.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase());
  };

  return (
    <Card padding="lg" className="relative overflow-hidden">
      {/* Status Badge */}
      <div className="flex items-center justify-between mb-4">
        <Badge variant={isExpired ? 'error' : 'info'}>
          {isExpired ? 'Expired' : 'Pending'}
        </Badge>
        <span className="text-sm text-[var(--text-muted)]">
          {new Date(verification.created_at).toLocaleTimeString()}
        </span>
      </div>

      {/* Operation Type */}
      <h3 className="text-lg font-semibold text-[var(--text-primary)] mb-2">
        {getOperationLabel(verification.operation_type)}
      </h3>

      {/* OTP Code Display */}
      {verification.otp_code && !isExpired ? (
        <div className="my-6">
          <p className="text-sm text-[var(--text-secondary)] mb-2">Verification Code:</p>
          <div
            onClick={handleCopy}
            className="flex items-center justify-center gap-4 p-4 bg-[var(--surface-elevated)] rounded-xl cursor-pointer hover:bg-[var(--color-primary-50)] transition-colors"
            role="button"
            tabIndex={0}
            onKeyDown={e => e.key === 'Enter' && handleCopy()}
            aria-label={`Copy OTP code ${verification.otp_code}`}
          >
            <span className="font-mono text-4xl tracking-[0.3em] font-bold text-[var(--color-primary-600)]">
              {verification.otp_code}
            </span>
            <button
              className="p-2 rounded-lg bg-[var(--surface-card)] border border-[var(--border-default)] hover:bg-[var(--color-primary-50)]"
              title="Copy to clipboard"
            >
              {copied ? (
                <svg className="w-5 h-5 text-[var(--text-success)]" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                </svg>
              ) : (
                <svg className="w-5 h-5 text-[var(--text-muted)]" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
                </svg>
              )}
            </button>
          </div>
          {copied && (
            <p className="text-center text-sm text-[var(--text-success)] mt-2">
              Copied to clipboard!
            </p>
          )}
        </div>
      ) : (
        <div className="my-6 p-4 bg-[var(--surface-error)] rounded-xl text-center">
          <p className="text-[var(--text-error)] font-medium">
            {isExpired ? 'This verification has expired' : 'OTP code not available'}
          </p>
        </div>
      )}

      {/* Timer */}
      {!isExpired && (
        <div className={`flex items-center justify-center gap-2 p-3 rounded-lg ${isUrgent ? 'bg-[var(--surface-warning)]' : 'bg-[var(--surface-page)]'}`}>
          <svg className={`w-5 h-5 ${isUrgent ? 'text-[var(--text-warning)]' : 'text-[var(--text-muted)]'}`} fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <span className={`font-mono text-lg font-semibold ${isUrgent ? 'text-[var(--text-warning)]' : 'text-[var(--text-secondary)]'}`}>
            {formatTime(timeRemaining)}
          </span>
          <span className={`text-sm ${isUrgent ? 'text-[var(--text-warning)]' : 'text-[var(--text-muted)]'}`}>
            remaining
          </span>
        </div>
      )}

      {/* Metadata */}
      {verification.metadata && Object.keys(verification.metadata).length > 0 && (
        <div className="mt-4 pt-4 border-t border-[var(--border-subtle)]">
          <p className="text-xs text-[var(--text-muted)] mb-2">Details:</p>
          <div className="text-sm text-[var(--text-secondary)]">
            {Object.entries(verification.metadata).map(([key, value]) => (
              <p key={key}>
                <span className="text-[var(--text-muted)]">{key.replace(/_/g, ' ')}:</span>{' '}
                {String(value)}
              </p>
            ))}
          </div>
        </div>
      )}
    </Card>
  );
}
