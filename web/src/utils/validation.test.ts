import { describe, it, expect } from 'vitest';
import {
  validateEmail,
  validatePassword,
  checkPasswordStrength,
  validateName,
  validateDescription,
  validateUUID,
  validateURL,
  sanitizeString,
  sanitizeHTML,
  sanitizeSearchQuery,
  containsXSSPattern,
  validateForm,
  type ValidationRule,
} from './validation';

describe('validateEmail', () => {
  it('should accept valid emails', () => {
    expect(validateEmail('test@example.com').valid).toBe(true);
    expect(validateEmail('test.user@example.com').valid).toBe(true);
    expect(validateEmail('test+label@example.com').valid).toBe(true);
    expect(validateEmail('test@subdomain.example.com').valid).toBe(true);
  });

  it('should reject empty email', () => {
    const result = validateEmail('');
    expect(result.valid).toBe(false);
    expect(result.errors[0].code).toBe('required');
  });

  it('should reject invalid emails', () => {
    expect(validateEmail('notanemail').valid).toBe(false);
    expect(validateEmail('test@').valid).toBe(false);
    expect(validateEmail('@example.com').valid).toBe(false);
    expect(validateEmail('test@example').valid).toBe(false);
  });

  it('should reject too long emails', () => {
    const longEmail = 'a'.repeat(250) + '@example.com';
    const result = validateEmail(longEmail);
    expect(result.valid).toBe(false);
    expect(result.errors[0].code).toBe('max_length');
  });
});

describe('validatePassword', () => {
  it('should accept strong passwords', () => {
    expect(validatePassword('MyStr0ng!Pass#123').valid).toBe(true);
    expect(validatePassword('Password123!').valid).toBe(true);
  });

  it('should reject empty password', () => {
    const result = validatePassword('');
    expect(result.valid).toBe(false);
    expect(result.errors[0].code).toBe('required');
  });

  it('should reject short passwords', () => {
    const result = validatePassword('Short1!');
    expect(result.valid).toBe(false);
    expect(result.errors[0].code).toBe('min_length');
  });

  it('should reject weak passwords', () => {
    expect(validatePassword('alllowercase1').valid).toBe(false);
    expect(validatePassword('ALLUPPERCASE1').valid).toBe(false);
    expect(validatePassword('123456789012').valid).toBe(false);
  });

  it('should reject too long passwords', () => {
    const longPassword = 'Aa1!'.repeat(50);
    const result = validatePassword(longPassword);
    expect(result.valid).toBe(false);
    expect(result.errors[0].code).toBe('max_length');
  });
});

describe('checkPasswordStrength', () => {
  it('should return weak for empty password', () => {
    expect(checkPasswordStrength('').score).toBe(0);
    expect(checkPasswordStrength('').label).toBe('Weak');
  });

  it('should return weak for short passwords', () => {
    expect(checkPasswordStrength('abc').score).toBeLessThan(2);
  });

  it('should return strong for complex passwords', () => {
    const result = checkPasswordStrength('MyStr0ng!Password#123');
    expect(result.score).toBeGreaterThanOrEqual(3);
    expect(result.label).toBe('Strong');
  });

  it('should provide feedback for weak passwords', () => {
    const result = checkPasswordStrength('password');
    expect(result.feedback.length).toBeGreaterThan(0);
  });
});

describe('validateName', () => {
  it('should accept valid names', () => {
    expect(validateName('My Workflow').valid).toBe(true);
    expect(validateName('Workflow 123').valid).toBe(true);
    expect(validateName('My_Workflow-Test').valid).toBe(true);
    expect(validateName("John's Workflow").valid).toBe(true);
  });

  it('should reject empty names', () => {
    const result = validateName('');
    expect(result.valid).toBe(false);
    expect(result.errors[0].code).toBe('required');
  });

  it('should reject names with null bytes', () => {
    const result = validateName('Test\x00Name');
    expect(result.valid).toBe(false);
    expect(result.errors[0].code).toBe('invalid_chars');
  });

  it('should reject too long names', () => {
    const longName = 'a'.repeat(300);
    const result = validateName(longName);
    expect(result.valid).toBe(false);
    expect(result.errors[0].code).toBe('max_length');
  });
});

describe('validateDescription', () => {
  it('should accept empty descriptions', () => {
    expect(validateDescription('').valid).toBe(true);
  });

  it('should accept valid descriptions', () => {
    expect(validateDescription('This is a valid description.').valid).toBe(true);
  });

  it('should reject descriptions with null bytes', () => {
    const result = validateDescription('Test\x00Description');
    expect(result.valid).toBe(false);
    expect(result.errors[0].code).toBe('invalid_chars');
  });

  it('should reject too long descriptions', () => {
    const longDesc = 'a'.repeat(6000);
    const result = validateDescription(longDesc);
    expect(result.valid).toBe(false);
    expect(result.errors[0].code).toBe('max_length');
  });
});

describe('validateUUID', () => {
  it('should accept valid UUIDs', () => {
    expect(validateUUID('550e8400-e29b-41d4-a716-446655440000').valid).toBe(true);
    expect(validateUUID('550E8400-E29B-41D4-A716-446655440000').valid).toBe(true);
  });

  it('should reject empty UUIDs', () => {
    const result = validateUUID('');
    expect(result.valid).toBe(false);
    expect(result.errors[0].code).toBe('required');
  });

  it('should reject invalid UUIDs', () => {
    expect(validateUUID('not-a-uuid').valid).toBe(false);
    expect(validateUUID('550e8400e29b41d4a716446655440000').valid).toBe(false);
    expect(validateUUID('550e8400-e29b-41d4-a716').valid).toBe(false);
    expect(validateUUID('550e8400-e29b-41d4-a716-44665544000g').valid).toBe(false);
  });
});

describe('validateURL', () => {
  it('should accept valid URLs', () => {
    expect(validateURL('https://example.com').valid).toBe(true);
    expect(validateURL('http://example.com/path').valid).toBe(true);
    expect(validateURL('https://api.example.com/v1/endpoint').valid).toBe(true);
  });

  it('should reject empty URLs', () => {
    const result = validateURL('');
    expect(result.valid).toBe(false);
    expect(result.errors[0].code).toBe('required');
  });

  it('should reject invalid URLs', () => {
    expect(validateURL('not-a-url').valid).toBe(false);
  });

  it('should reject non-http protocols', () => {
    const result = validateURL('ftp://example.com');
    expect(result.valid).toBe(false);
    expect(result.errors[0].code).toBe('invalid_protocol');
  });

  it('should reject localhost URLs', () => {
    expect(validateURL('http://localhost').valid).toBe(false);
    expect(validateURL('http://127.0.0.1').valid).toBe(false);
    expect(validateURL('http://test.localhost').valid).toBe(false);
  });
});

describe('sanitizeString', () => {
  it('should handle empty strings', () => {
    expect(sanitizeString('')).toBe('');
  });

  it('should remove null bytes', () => {
    expect(sanitizeString('Hello\x00World')).toBe('HelloWorld');
  });

  it('should trim whitespace', () => {
    expect(sanitizeString('  Hello World  ')).toBe('Hello World');
  });
});

describe('sanitizeHTML', () => {
  it('should handle empty strings', () => {
    expect(sanitizeHTML('')).toBe('');
  });

  it('should escape HTML entities', () => {
    expect(sanitizeHTML('<script>alert("xss")</script>')).toBe(
      '&lt;script&gt;alert(&quot;xss&quot;)&lt;/script&gt;'
    );
  });

  it('should escape quotes', () => {
    expect(sanitizeHTML('"quoted" & \'single\'')).toBe(
      '&quot;quoted&quot; &amp; &#39;single&#39;'
    );
  });
});

describe('sanitizeSearchQuery', () => {
  it('should handle empty strings', () => {
    expect(sanitizeSearchQuery('')).toBe('');
  });

  it('should trim and limit length', () => {
    const longQuery = 'a'.repeat(600);
    expect(sanitizeSearchQuery(longQuery).length).toBe(500);
  });

  it('should remove null bytes', () => {
    expect(sanitizeSearchQuery('test\x00query')).toBe('testquery');
  });
});

describe('containsXSSPattern', () => {
  it('should return false for safe strings', () => {
    expect(containsXSSPattern('Hello World')).toBe(false);
    expect(containsXSSPattern('normal text')).toBe(false);
  });

  it('should detect script tags', () => {
    expect(containsXSSPattern('<script>alert("xss")</script>')).toBe(true);
  });

  it('should detect javascript protocol', () => {
    expect(containsXSSPattern('javascript:alert("xss")')).toBe(true);
  });

  it('should detect event handlers', () => {
    expect(containsXSSPattern('<img onerror="alert(1)">')).toBe(true);
    expect(containsXSSPattern('<div onclick="evil()">')).toBe(true);
  });

  it('should detect iframe tags', () => {
    expect(containsXSSPattern('<iframe src="evil.com">')).toBe(true);
  });
});

describe('validateForm', () => {
  interface TestForm {
    email: string;
    name: string;
  }

  const rules: ValidationRule<TestForm>[] = [
    {
      field: 'email',
      validate: (value) => validateEmail(value as string),
    },
    {
      field: 'name',
      validate: (value) => validateName(value as string),
    },
  ];

  it('should validate all fields', () => {
    const validData: TestForm = {
      email: 'test@example.com',
      name: 'Test Name',
    };
    expect(validateForm(validData, rules).valid).toBe(true);
  });

  it('should collect errors from all fields', () => {
    const invalidData: TestForm = {
      email: 'invalid',
      name: '',
    };
    const result = validateForm(invalidData, rules);
    expect(result.valid).toBe(false);
    expect(result.errors.length).toBe(2);
  });
});
