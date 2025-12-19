/**
 * Converts a cron expression to a human-readable description
 * @param expression - The cron expression (e.g., "0 9 * * 1-5")
 * @returns A human-readable description
 */
export function describeCron(expression: string): string {
  if (!expression || typeof expression !== 'string') {
    return 'Custom schedule';
  }

  const parts = expression.trim().split(/\s+/);
  if (parts.length < 5) {
    return 'Custom schedule';
  }

  const [minute, hour, dayOfMonth, month, dayOfWeek] = parts;

  try {
    // Every X minutes
    if (minute.startsWith('*/') && hour === '*' && dayOfMonth === '*' && month === '*' && dayOfWeek === '*') {
      const interval = minute.substring(2);
      return `Every ${interval} minutes`;
    }

    // Every X hours
    if (minute === '0' && hour.startsWith('*/') && dayOfMonth === '*' && month === '*' && dayOfWeek === '*') {
      const interval = hour.substring(2);
      return `Every ${interval} hours`;
    }

    // Every hour
    if (minute === '0' && hour === '*' && dayOfMonth === '*' && month === '*' && dayOfWeek === '*') {
      return 'Every hour';
    }

    // Weekdays pattern
    if (dayOfWeek === '1-5' && dayOfMonth === '*' && month === '*') {
      const time = formatTime(hour, minute);
      return `Weekdays at ${time}`;
    }

    // Daily at specific time
    if (dayOfMonth === '*' && month === '*' && dayOfWeek === '*') {
      const time = formatTime(hour, minute);

      // Check for multiple hours
      if (hour.includes(',')) {
        const hours = hour.split(',');
        const times = hours.map(h => formatTime(h, minute));
        return `At ${times.join(', ')} daily`;
      }

      return `Daily at ${time}`;
    }

    // Specific day of week
    if (dayOfWeek.match(/^\d$/) && dayOfMonth === '*' && month === '*') {
      const dayName = getDayName(dayOfWeek);
      const time = formatTime(hour, minute);
      return `Every ${dayName} at ${time}`;
    }

    // Specific day of month
    if (dayOfMonth.match(/^\d+$/) && month === '*' && dayOfWeek === '*') {
      const time = formatTime(hour, minute);
      return `Monthly on day ${dayOfMonth} at ${time}`;
    }

    // Quarterly or specific months
    if (month.includes(',') && dayOfMonth.match(/^\d+$/)) {
      const monthNames = month.split(',').map(m => getMonthName(m));
      const time = formatTime(hour, minute);
      return `On day ${dayOfMonth} of ${monthNames.join(', ')} at ${time}`;
    }

    // Complex day of week with day ranges (bi-weekly pattern)
    if (dayOfWeek.match(/^\d$/) && dayOfMonth.includes(',')) {
      const dayName = getDayName(dayOfWeek);
      const time = formatTime(hour, minute);
      const formattedDays = dayOfMonth.replace(/,/g, ', ');
      return `Every ${dayName} during days ${formattedDays} at ${time}`;
    }

    // Fallback for unrecognized patterns
    return 'Custom schedule';
  } catch {
    return 'Custom schedule';
  }
}

/**
 * Formats hour and minute into readable time (e.g., "9:00 AM")
 */
function formatTime(hour: string, minute: string): string {
  const h = parseInt(hour, 10);
  const m = parseInt(minute, 10);

  if (isNaN(h) || isNaN(m)) {
    return `${hour}:${minute}`;
  }

  const period = h >= 12 ? 'PM' : 'AM';
  const hour12 = h === 0 ? 12 : h > 12 ? h - 12 : h;
  const minuteStr = m.toString().padStart(2, '0');

  return `${hour12}:${minuteStr} ${period}`;
}

/**
 * Gets day name from day of week number (0 = Sunday, 6 = Saturday)
 */
function getDayName(dayOfWeek: string): string {
  const days = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday'];
  const day = parseInt(dayOfWeek, 10);
  return days[day] || dayOfWeek;
}

/**
 * Gets month name from month number (1 = January, 12 = December)
 */
function getMonthName(month: string): string {
  const months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];
  const m = parseInt(month, 10) - 1;
  return months[m] || month;
}
