/**
 * Badge Variant Helpers
 * Centralized mapping of status values to badge variants
 */

export type BadgeVariant = 'success' | 'warning' | 'error' | 'info' | 'neutral';

/**
 * Get badge variant for user/wallet status
 */
export function getStatusVariant(status: string): BadgeVariant {
  switch (status) {
    case 'active':
      return 'success';
    case 'pending':
      return 'warning';
    case 'suspended':
    case 'frozen':
    case 'closed':
      return 'error';
    default:
      return 'neutral';
  }
}

/**
 * Get badge variant for KYC status
 */
export function getKYCStatusVariant(status?: string): BadgeVariant {
  switch (status) {
    case 'verified':
    case 'approved':
      return 'success';
    case 'pending':
    case 'submitted':
      return 'warning';
    case 'rejected':
      return 'error';
    default:
      return 'neutral';
  }
}

/**
 * Get badge variant for transaction status
 */
export function getTransactionStatusVariant(status: string): BadgeVariant {
  switch (status) {
    case 'completed':
      return 'success';
    case 'pending':
    case 'processing':
      return 'warning';
    case 'failed':
      return 'error';
    case 'reversed':
    case 'cancelled':
      return 'info';
    default:
      return 'neutral';
  }
}

/**
 * Get badge variant for transaction type
 */
export function getTransactionTypeVariant(type: string): BadgeVariant {
  switch (type) {
    case 'deposit':
    case 'refund':
      return 'success';
    case 'withdrawal':
      return 'warning';
    case 'transfer':
    case 'reversal':
      return 'info';
    case 'fee':
      return 'error';
    default:
      return 'neutral';
  }
}

/**
 * Get badge variant for wallet status
 */
export function getWalletStatusVariant(status: string): BadgeVariant {
  switch (status) {
    case 'active':
      return 'success';
    case 'frozen':
      return 'warning';
    case 'closed':
      return 'error';
    default:
      return 'neutral';
  }
}
