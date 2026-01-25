/**
 * Input Validation Utilities
 *
 * Client-side validation that mirrors backend validation rules.
 * These provide immediate feedback to users but should NEVER replace
 * server-side validation.
 */

export interface ValidationError {
  field: string;
  message: string;
  code: string;
}

export interface ValidationResult {
  valid: boolean;
  errors: ValidationError[];
}

// --- Email Validation ---

const EMAIL_REGEX = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/;
const MAX_EMAIL_LENGTH = 254;

export function validateEmail(email: string, fieldName = 'email'): ValidationResult {
  const errors: ValidationError[] = [];

  if (!email || email.trim() === '') {
    errors.push({
      field: fieldName,
      message: `${fieldName} is required`,
      code: 'required',
    });
    return { valid: false, errors };
  }

  if (email.length > MAX_EMAIL_LENGTH) {
    errors.push({
      field: fieldName,
      message: `${fieldName} exceeds maximum length of ${MAX_EMAIL_LENGTH} characters`,
      code: 'max_length',
    });
    return { valid: false, errors };
  }

  if (!EMAIL_REGEX.test(email)) {
    errors.push({
      field: fieldName,
      message: 'Invalid email format',
      code: 'invalid_format',
    });
  }

  return { valid: errors.length === 0, errors };
}

// --- Password Validation ---

const MIN_PASSWORD_LENGTH = 12;
const MAX_PASSWORD_LENGTH = 128;

export interface PasswordStrength {
  score: number; // 0-4
  label: 'Weak' | 'Fair' | 'Good' | 'Strong';
  feedback: string[];
}

export function validatePassword(
  password: string,
  fieldName = 'password'
): ValidationResult {
  const errors: ValidationError[] = [];

  if (!password) {
    errors.push({
      field: fieldName,
      message: `${fieldName} is required`,
      code: 'required',
    });
    return { valid: false, errors };
  }

  if (password.length < MIN_PASSWORD_LENGTH) {
    errors.push({
      field: fieldName,
      message: `${fieldName} must be at least ${MIN_PASSWORD_LENGTH} characters`,
      code: 'min_length',
    });
  }

  if (password.length > MAX_PASSWORD_LENGTH) {
    errors.push({
      field: fieldName,
      message: `${fieldName} must not exceed ${MAX_PASSWORD_LENGTH} characters`,
      code: 'max_length',
    });
  }

  // Check character variety
  const hasUpper = /[A-Z]/.test(password);
  const hasLower = /[a-z]/.test(password);
  const hasDigit = /[0-9]/.test(password);
  const hasSpecial = /[!@#$%^&*()_+\-=[\]{};':"\\|,.<>/?]/.test(password);

  const typeCount = [hasUpper, hasLower, hasDigit, hasSpecial].filter(Boolean).length;

  if (typeCount < 3) {
    errors.push({
      field: fieldName,
      message: 'Password must contain at least 3 of: uppercase, lowercase, digit, special character',
      code: 'weak_password',
    });
  }

  return { valid: errors.length === 0, errors };
}

export function checkPasswordStrength(password: string): PasswordStrength {
  const feedback: string[] = [];
  let score = 0;

  if (!password) {
    return { score: 0, label: 'Weak', feedback: ['Enter a password'] };
  }

  // Length checks
  if (password.length >= MIN_PASSWORD_LENGTH) score++;
  if (password.length >= 16) score++;
  if (password.length < MIN_PASSWORD_LENGTH) {
    feedback.push(`Use at least ${MIN_PASSWORD_LENGTH} characters`);
  }

  // Character variety
  const hasUpper = /[A-Z]/.test(password);
  const hasLower = /[a-z]/.test(password);
  const hasDigit = /[0-9]/.test(password);
  const hasSpecial = /[!@#$%^&*()_+\-=[\]{};':"\\|,.<>/?]/.test(password);

  if (hasUpper && hasLower) score++;
  if (hasDigit && hasSpecial) score++;

  if (!hasUpper) feedback.push('Add uppercase letters');
  if (!hasLower) feedback.push('Add lowercase letters');
  if (!hasDigit) feedback.push('Add numbers');
  if (!hasSpecial) feedback.push('Add special characters');

  // Cap at 4
  score = Math.min(score, 4);

  const labels: PasswordStrength['label'][] = ['Weak', 'Fair', 'Good', 'Strong', 'Strong'];
  return { score, label: labels[score], feedback };
}

// --- Name/String Validation ---

const MAX_NAME_LENGTH = 255;
const MAX_DESCRIPTION_LENGTH = 5000;
// Allows alphanumeric, spaces, hyphens, underscores, and common punctuation
const NAME_REGEX = /^[\p{L}\p{N}\s\-_.,!?'":;()&]+$/u;

export function validateName(
  name: string,
  fieldName = 'name',
  maxLength = MAX_NAME_LENGTH
): ValidationResult {
  const errors: ValidationError[] = [];

  if (!name || name.trim() === '') {
    errors.push({
      field: fieldName,
      message: `${fieldName} is required`,
      code: 'required',
    });
    return { valid: false, errors };
  }

  if (name.length > maxLength) {
    errors.push({
      field: fieldName,
      message: `${fieldName} must not exceed ${maxLength} characters`,
      code: 'max_length',
    });
  }

  // Check for null bytes
  if (name.includes('\x00')) {
    errors.push({
      field: fieldName,
      message: `${fieldName} contains invalid characters`,
      code: 'invalid_chars',
    });
  }

  // Validate allowed characters
  if (!NAME_REGEX.test(name)) {
    errors.push({
      field: fieldName,
      message: `${fieldName} contains invalid characters`,
      code: 'invalid_chars',
    });
  }

  return { valid: errors.length === 0, errors };
}

export function validateDescription(
  description: string,
  fieldName = 'description'
): ValidationResult {
  const errors: ValidationError[] = [];

  // Descriptions can be empty
  if (!description || description.trim() === '') {
    return { valid: true, errors };
  }

  if (description.length > MAX_DESCRIPTION_LENGTH) {
    errors.push({
      field: fieldName,
      message: `${fieldName} must not exceed ${MAX_DESCRIPTION_LENGTH} characters`,
      code: 'max_length',
    });
  }

  // Check for null bytes
  if (description.includes('\x00')) {
    errors.push({
      field: fieldName,
      message: `${fieldName} contains invalid characters`,
      code: 'invalid_chars',
    });
  }

  return { valid: errors.length === 0, errors };
}

// --- UUID Validation ---

const UUID_REGEX =
  /^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$/;

export function validateUUID(id: string, fieldName = 'id'): ValidationResult {
  const errors: ValidationError[] = [];

  if (!id) {
    errors.push({
      field: fieldName,
      message: `${fieldName} is required`,
      code: 'required',
    });
    return { valid: false, errors };
  }

  if (!UUID_REGEX.test(id)) {
    errors.push({
      field: fieldName,
      message: `${fieldName} must be a valid UUID`,
      code: 'invalid_format',
    });
  }

  return { valid: errors.length === 0, errors };
}

// --- Sanitization Functions ---

/**
 * Remove or escape potentially dangerous characters
 */
export function sanitizeString(s: string): string {
  if (!s) return '';

  // Remove null bytes
  let sanitized = s.replace(/\x00/g, '');

  // Trim whitespace
  sanitized = sanitized.trim();

  return sanitized;
}

/**
 * Escape HTML entities to prevent XSS
 */
export function sanitizeHTML(s: string): string {
  if (!s) return '';

  const htmlEscapes: Record<string, string> = {
    '&': '&amp;',
    '<': '&lt;',
    '>': '&gt;',
    '"': '&quot;',
    "'": '&#39;',
  };

  return s.replace(/[&<>"']/g, (char) => htmlEscapes[char] || char);
}

/**
 * Sanitize search query input
 */
export function sanitizeSearchQuery(query: string): string {
  if (!query) return '';

  // Remove null bytes
  let sanitized = query.replace(/\x00/g, '');

  // Trim and limit length
  sanitized = sanitized.trim();
  if (sanitized.length > 500) {
    sanitized = sanitized.substring(0, 500);
  }

  return sanitized;
}

// --- XSS Pattern Detection ---

const XSS_PATTERNS = [
  '<script',
  '</script',
  'javascript:',
  'vbscript:',
  'onload=',
  'onerror=',
  'onclick=',
  'onmouseover=',
  'onfocus=',
  'onblur=',
  '<iframe',
  '<object',
  '<embed',
  '<svg',
  '<math',
  'expression(',
  'url(',
  'data:',
];

/**
 * Check if a string contains potential XSS patterns
 */
export function containsXSSPattern(s: string): boolean {
  if (!s) return false;
  const lower = s.toLowerCase();
  return XSS_PATTERNS.some((pattern) => lower.includes(pattern));
}

// --- Form Validation Helper ---

export type ValidationRule<T> = {
  field: keyof T;
  validate: (value: unknown, data: T) => ValidationResult;
};

export function validateForm<T extends Record<string, unknown>>(
  data: T,
  rules: ValidationRule<T>[]
): ValidationResult {
  const allErrors: ValidationError[] = [];

  for (const rule of rules) {
    const value = data[rule.field];
    const result = rule.validate(value, data);
    if (!result.valid) {
      allErrors.push(...result.errors);
    }
  }

  return { valid: allErrors.length === 0, errors: allErrors };
}

// --- URL Validation ---

const ALLOWED_PROTOCOLS = ['http:', 'https:'];

export function validateURL(url: string, fieldName = 'url'): ValidationResult {
  const errors: ValidationError[] = [];

  if (!url || url.trim() === '') {
    errors.push({
      field: fieldName,
      message: `${fieldName} is required`,
      code: 'required',
    });
    return { valid: false, errors };
  }

  try {
    const parsed = new URL(url);

    if (!ALLOWED_PROTOCOLS.includes(parsed.protocol)) {
      errors.push({
        field: fieldName,
        message: 'Only http and https protocols are allowed',
        code: 'invalid_protocol',
      });
    }

    // Check for localhost in production
    if (
      parsed.hostname === 'localhost' ||
      parsed.hostname === '127.0.0.1' ||
      parsed.hostname.endsWith('.localhost')
    ) {
      errors.push({
        field: fieldName,
        message: 'Local URLs are not allowed',
        code: 'local_url',
      });
    }
  } catch {
    errors.push({
      field: fieldName,
      message: 'Invalid URL format',
      code: 'invalid_format',
    });
  }

  return { valid: errors.length === 0, errors };
}
