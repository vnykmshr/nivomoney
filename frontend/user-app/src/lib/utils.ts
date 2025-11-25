/**
 * Format amount in paise to INR currency
 */
export function formatCurrency(amountPaise: number): string {
  const amount = amountPaise / 100;
  return new Intl.NumberFormat('en-IN', {
    style: 'currency',
    currency: 'INR',
    minimumFractionDigits: 2,
  }).format(amount);
}

/**
 * Convert INR amount to paise
 */
export function toPaise(amount: number): number {
  return Math.round(amount * 100);
}

/**
 * Format date to human-readable format
 */
export function formatDate(dateString: string): string {
  const date = new Date(dateString);
  return new Intl.DateTimeFormat('en-IN', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  }).format(date);
}

/**
 * Format date to relative time (e.g., "2 hours ago")
 */
export function formatRelativeTime(dateString: string): string {
  const date = new Date(dateString);
  const now = new Date();
  const diffInSeconds = Math.floor((now.getTime() - date.getTime()) / 1000);

  if (diffInSeconds < 60) return 'Just now';
  if (diffInSeconds < 3600) return `${Math.floor(diffInSeconds / 60)} min ago`;
  if (diffInSeconds < 86400) return `${Math.floor(diffInSeconds / 3600)} hours ago`;
  if (diffInSeconds < 604800) return `${Math.floor(diffInSeconds / 86400)} days ago`;

  return formatDate(dateString);
}

/**
 * Get badge color based on status
 */
export function getStatusColor(status: string): string {
  const colors: Record<string, string> = {
    pending: 'bg-yellow-100 text-yellow-800',
    processing: 'bg-blue-100 text-blue-800',
    completed: 'bg-green-100 text-green-800',
    failed: 'bg-red-100 text-red-800',
    reversed: 'bg-gray-100 text-gray-800',
    active: 'bg-green-100 text-green-800',
    inactive: 'bg-gray-100 text-gray-800',
    frozen: 'bg-blue-100 text-blue-800',
    closed: 'bg-red-100 text-red-800',
    verified: 'bg-green-100 text-green-800',
    rejected: 'bg-red-100 text-red-800',
  };

  return colors[status] || 'bg-gray-100 text-gray-800';
}

/**
 * Validate email format
 */
export function isValidEmail(email: string): boolean {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  return emailRegex.test(email);
}

/**
 * Validate phone number (Indian format)
 */
export function isValidPhone(phone: string): boolean {
  const phoneRegex = /^[6-9]\d{9}$/;
  return phoneRegex.test(phone.replace(/\s/g, ''));
}

/**
 * Normalize Indian phone number by adding +91 prefix if needed
 * Accepts: 10-digit number (9876543210) -> Returns: +919876543210
 * Already formatted (+919876543210) -> Returns as-is
 * Email or other -> Returns as-is
 */
export function normalizeIndianPhone(input: string): string {
  // Remove any whitespace
  const trimmed = input.trim().replace(/\s/g, '');

  // Pattern: exactly 10 digits starting with 6, 7, 8, or 9
  const tenDigitPattern = /^[6-9][0-9]{9}$/;

  if (tenDigitPattern.test(trimmed)) {
    return '+91' + trimmed;
  }

  return trimmed;
}
