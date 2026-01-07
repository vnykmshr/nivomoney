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
