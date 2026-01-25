import { describe, it, expect } from 'vitest'
import { isValidResourceId, validateResourceIds, getValidatedId } from './routing'

describe('routing utilities', () => {
  describe('isValidResourceId', () => {
    it('should return true for valid UUIDs', () => {
      const validUUIDs = [
        '550e8400-e29b-41d4-a716-446655440000',
        'f47ac10b-58cc-4372-a567-0e02b2c3d479',
        '123e4567-e89b-12d3-a456-426614174000',
      ]

      validUUIDs.forEach((uuid) => {
        expect(isValidResourceId(uuid)).toBe(true)
      })
    })

    it('should return false for reserved route keywords', () => {
      const reservedKeywords = ['new', 'create', 'edit', 'list', 'settings', 'NEW', 'Create', 'EDIT']

      reservedKeywords.forEach((keyword) => {
        expect(isValidResourceId(keyword)).toBe(false)
      })
    })

    it('should return false for invalid UUIDs', () => {
      const invalidUUIDs = [
        'not-a-uuid',
        '12345',
        'abc',
        '550e8400-e29b-41d4-a716',
        '550e8400-e29b-41d4-a716-44665544000',
        'xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx',
        '',
      ]

      invalidUUIDs.forEach((uuid) => {
        expect(isValidResourceId(uuid)).toBe(false)
      })
    })

    it('should return false for undefined', () => {
      expect(isValidResourceId(undefined)).toBe(false)
    })

    it('should return false for null', () => {
      expect(isValidResourceId(null as any)).toBe(false)
    })

    it('should return false for empty string', () => {
      expect(isValidResourceId('')).toBe(false)
    })
  })

  describe('validateResourceIds', () => {
    it('should validate multiple IDs correctly', () => {
      const ids = {
        validId: '550e8400-e29b-41d4-a716-446655440000',
        invalidId: 'new',
        anotherId: 'f47ac10b-58cc-4372-a567-0e02b2c3d479',
        emptyId: '',
      }

      const results = validateResourceIds(ids)

      expect(results).toEqual({
        validId: true,
        invalidId: false,
        anotherId: true,
        emptyId: false,
      })
    })

    it('should handle empty object', () => {
      const results = validateResourceIds({})
      expect(results).toEqual({})
    })

    it('should handle undefined values', () => {
      const ids = {
        id1: undefined,
        id2: '550e8400-e29b-41d4-a716-446655440000',
      }

      const results = validateResourceIds(ids)

      expect(results).toEqual({
        id1: false,
        id2: true,
      })
    })
  })

  describe('getValidatedId', () => {
    it('should extract valid ID from URLSearchParams', () => {
      const params = new URLSearchParams('id=550e8400-e29b-41d4-a716-446655440000')
      expect(getValidatedId(params)).toBe('550e8400-e29b-41d4-a716-446655440000')
    })

    it('should return null for invalid ID in URLSearchParams', () => {
      const params = new URLSearchParams('id=new')
      expect(getValidatedId(params)).toBe(null)
    })

    it('should extract valid ID from Record', () => {
      const params = { id: '550e8400-e29b-41d4-a716-446655440000' }
      expect(getValidatedId(params)).toBe('550e8400-e29b-41d4-a716-446655440000')
    })

    it('should return null for invalid ID in Record', () => {
      const params = { id: 'create' }
      expect(getValidatedId(params)).toBe(null)
    })

    it('should handle missing parameter', () => {
      const params = new URLSearchParams('')
      expect(getValidatedId(params)).toBe(null)
    })

    it('should handle custom parameter name', () => {
      const params = { customId: '550e8400-e29b-41d4-a716-446655440000' }
      expect(getValidatedId(params, 'customId')).toBe('550e8400-e29b-41d4-a716-446655440000')
    })
  })
})
