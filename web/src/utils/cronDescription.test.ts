import { describe, it, expect } from 'vitest';
import { describeCron } from './cronDescription';

describe('describeCron', () => {
  describe('hourly patterns', () => {
    it('should describe every hour', () => {
      expect(describeCron('0 * * * *')).toBe('Every hour');
    });

    it('should describe every 2 hours', () => {
      expect(describeCron('0 */2 * * *')).toBe('Every 2 hours');
    });

    it('should describe every 4 hours', () => {
      expect(describeCron('0 */4 * * *')).toBe('Every 4 hours');
    });

    it('should describe every 6 hours', () => {
      expect(describeCron('0 */6 * * *')).toBe('Every 6 hours');
    });
  });

  describe('minute patterns', () => {
    it('should describe every 15 minutes', () => {
      expect(describeCron('*/15 * * * *')).toBe('Every 15 minutes');
    });

    it('should describe every 30 minutes', () => {
      expect(describeCron('*/30 * * * *')).toBe('Every 30 minutes');
    });

    it('should describe every 5 minutes', () => {
      expect(describeCron('*/5 * * * *')).toBe('Every 5 minutes');
    });
  });

  describe('daily patterns', () => {
    it('should describe daily at midnight', () => {
      expect(describeCron('0 0 * * *')).toBe('Daily at 12:00 AM');
    });

    it('should describe daily at 9 AM', () => {
      expect(describeCron('0 9 * * *')).toBe('Daily at 9:00 AM');
    });

    it('should describe daily at noon', () => {
      expect(describeCron('0 12 * * *')).toBe('Daily at 12:00 PM');
    });

    it('should describe daily at 6 PM', () => {
      expect(describeCron('0 18 * * *')).toBe('Daily at 6:00 PM');
    });

    it('should describe daily at 2 AM', () => {
      expect(describeCron('0 2 * * *')).toBe('Daily at 2:00 AM');
    });
  });

  describe('weekly patterns', () => {
    it('should describe weekdays at 9 AM', () => {
      expect(describeCron('0 9 * * 1-5')).toBe('Weekdays at 9:00 AM');
    });

    it('should describe Monday at 9 AM', () => {
      expect(describeCron('0 9 * * 1')).toBe('Every Monday at 9:00 AM');
    });

    it('should describe Friday at 5 PM', () => {
      expect(describeCron('0 17 * * 5')).toBe('Every Friday at 5:00 PM');
    });

    it('should describe Sunday at midnight', () => {
      expect(describeCron('0 0 * * 0')).toBe('Every Sunday at 12:00 AM');
    });

    it('should describe Saturday at 2 AM', () => {
      expect(describeCron('0 2 * * 6')).toBe('Every Saturday at 2:00 AM');
    });
  });

  describe('monthly patterns', () => {
    it('should describe monthly on 1st', () => {
      expect(describeCron('0 0 1 * *')).toBe('Monthly on day 1 at 12:00 AM');
    });

    it('should describe monthly on 15th at noon', () => {
      expect(describeCron('0 12 15 * *')).toBe('Monthly on day 15 at 12:00 PM');
    });

    it('should describe quarterly', () => {
      expect(describeCron('0 0 1 1,4,7,10 *')).toBe('On day 1 of Jan, Apr, Jul, Oct at 12:00 AM');
    });
  });

  describe('complex patterns', () => {
    it('should describe bi-weekly on Monday', () => {
      expect(describeCron('0 9 1-7,15-21 * 1')).toBe('Every Monday during days 1-7, 15-21 at 9:00 AM');
    });

    it('should describe weekdays at 8 AM', () => {
      expect(describeCron('0 8 * * 1-5')).toBe('Weekdays at 8:00 AM');
    });

    it('should describe weekdays at 6 PM', () => {
      expect(describeCron('0 18 * * 1-5')).toBe('Weekdays at 6:00 PM');
    });
  });

  describe('edge cases', () => {
    it('should handle invalid cron expression', () => {
      expect(describeCron('invalid')).toBe('Custom schedule');
    });

    it('should handle empty string', () => {
      expect(describeCron('')).toBe('Custom schedule');
    });

    it('should handle complex expression', () => {
      expect(describeCron('15 2,6,10 * * *')).toBe('At 2:15 AM, 6:15 AM, 10:15 AM daily');
    });
  });
});
