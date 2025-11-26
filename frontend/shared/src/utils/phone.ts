/**
 * Phone Number Utilities
 * Indian phone number normalization and validation
 */

/**
 * Normalize Indian phone number by adding +91 prefix if needed
 * @param input - Phone number or email
 * @returns Normalized phone number with +91 prefix, or input as-is if not a phone
 *
 * Examples:
 * - "9876543210" -> "+919876543210"
 * - "+919876543210" -> "+919876543210"
 * - "email@example.com" -> "email@example.com"
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

/**
 * Extract digits from phone number
 * @param phone - Phone number (any format)
 * @returns Only the digits
 */
export function extractPhoneDigits(phone: string): string {
  return phone.replace(/\D/g, '');
}

/**
 * Check if input is a phone number (not email)
 * @param input - User input (phone or email)
 * @returns true if it looks like a phone number
 */
export function isPhoneNumber(input: string): boolean {
  const digits = extractPhoneDigits(input);
  return /^[6-9][0-9]{9}$/.test(digits);
}

/**
 * Format phone for display with country code
 * @param phone - Phone number
 * @returns Formatted phone (e.g., "+91 98765 43210")
 */
export function formatIndianPhone(phone: string): string {
  const digits = extractPhoneDigits(phone);

  // If starts with 91 and is 12 digits
  if (digits.startsWith('91') && digits.length === 12) {
    const number = digits.substring(2);
    return `+91 ${number.substring(0, 5)} ${number.substring(5)}`;
  }

  // If 10 digits
  if (digits.length === 10 && /^[6-9]/.test(digits)) {
    return `+91 ${digits.substring(0, 5)} ${digits.substring(5)}`;
  }

  return phone;
}
