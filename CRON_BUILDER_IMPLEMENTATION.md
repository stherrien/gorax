# Visual Cron Expression Builder Implementation

## Overview
Implementation of a visual cron expression builder for the Gorax workflow scheduler, following TDD principles and clean code practices.

## Components Created

### Backend (Go)

#### 1. Schedule Service Enhancement (`internal/schedule/service.go`)
- **Method**: `GetNextRunTimes(expression, timezone string, count int) ([]time.Time, error)`
- Returns the next N execution times for a cron expression
- Validates cron expression and timezone
- Returns empty array for count <= 0
- **Tests**: `internal/schedule/service_test.go` - 8 comprehensive test cases

#### 2. API Handler (`internal/api/handlers/schedule.go`)
- **Endpoint**: `POST /api/v1/schedules/preview`
- Request body:
  ```json
  {
    "cron_expression": "0 9 * * *",
    "timezone": "UTC",
    "count": 10
  }
  ```
- Response:
  ```json
  {
    "valid": true,
    "next_runs": ["2025-12-20T09:00:00Z", ...],
    "count": 10,
    "timezone": "UTC"
  }
  ```
- Default count: 10, max: 50

#### 3. Route Configuration (`internal/api/app.go`)
- Added route: `r.Post("/preview", a.scheduleHandler.PreviewSchedule)`

### Frontend (React + TypeScript)

#### 1. API Client (`web/src/api/schedules.ts`)
- **Method**: `preview(cronExpression, timezone, count)`
- Added `PreviewScheduleResponse` interface
- **Tests**: 11 tests all passing

#### 2. CronBuilder Component (`web/src/components/schedule/CronBuilder.tsx`)
**Features**:
- Two modes: Simple (field-based) and Advanced (raw input)
- Quick presets dropdown:
  - Every minute
  - Every hour
  - Every day at midnight
  - Every day at 9 AM
  - Weekly on Monday at 9 AM
  - Monthly on the 1st
  - Weekdays at 9 AM
- Human-readable description of cron expression
- Real-time validation
- Visual error states

**Tests**: 13 tests all passing

#### 3. SchedulePreview Component (`web/src/components/schedule/SchedulePreview.tsx`)
**Features**:
- Shows next N execution times
- Displays timezone
- Formats dates with `date-fns` (e.g., "Dec 20, 2025 at 9:00 AM")
- Relative time hints ("tomorrow", "in 2 hours")
- Loading and error states
- Auto-refreshes when cron expression or timezone changes
- Scrollable list for many executions

**Tests**: 10 tests all passing

#### 4. TimezoneSelector Component (`web/src/components/schedule/TimezoneSelector.tsx`)
**Features**:
- Popular timezones at top (UTC, US Eastern, US Pacific, Europe/London)
- Grouped by region (America, Europe, Asia, Pacific, Africa)
- Shows timezone offset (e.g., "UTC-5")
- Optional search/filter capability
- Optional current time display
- 30+ timezones supported

**Tests**: 8 tests all passing

#### 5. Example Integration (`web/src/components/schedule/ScheduleBuilder.example.tsx`)
Demonstrates how to use all three components together in a real workflow scheduling form.

## Test Results

### Backend
- All `GetNextRunTimes` tests passing (8/8)
- Comprehensive coverage:
  - Valid expressions with multiple timezones
  - Edge cases (count=0, count=-1)
  - Invalid expressions
  - Different cron formats (@hourly, @daily, etc.)
  - Weekday-specific schedules

### Frontend
- **API Client**: 11/11 tests passing
- **CronBuilder**: 13/13 tests passing
- **SchedulePreview**: 10/10 tests passing
- **TimezoneSelector**: 8/8 tests passing
- **Total**: 42/42 tests passing

## Code Quality

### Backend
- ✅ Functions under 50 lines
- ✅ Cognitive complexity < 15
- ✅ Proper error handling with wrapped errors
- ✅ Clean separation of concerns
- ✅ Comprehensive test coverage

### Frontend
- ✅ Functional components only
- ✅ Custom hooks for reusable logic
- ✅ Proper TypeScript typing (no `any`)
- ✅ Responsive design with Tailwind CSS
- ✅ Accessible (ARIA labels, keyboard navigation)
- ✅ Loading and error states
- ✅ Clean component architecture

## Design Patterns Used

1. **Repository Pattern**: Data access abstraction in schedule service
2. **Service Layer Pattern**: Business logic separation
3. **Component Composition**: Building complex UIs from simple components
4. **Custom Hooks Pattern**: Reusable stateful logic (in SchedulePreview)
5. **Controlled Components**: React form pattern for all inputs

## TDD Approach

All code was developed following TDD:
1. ✅ **Red**: Wrote failing tests first
2. ✅ **Green**: Implemented minimal code to pass tests
3. ✅ **Refactor**: Cleaned up while maintaining green tests

## Usage Example

```typescript
import { CronBuilder, TimezoneSelector, SchedulePreview } from '@/components/schedule'

function ScheduleForm() {
  const [cron, setCron] = useState('0 9 * * *')
  const [tz, setTz] = useState('UTC')

  return (
    <div>
      <CronBuilder value={cron} onChange={setCron} />
      <TimezoneSelector value={tz} onChange={setTz} searchable />
      <SchedulePreview cronExpression={cron} timezone={tz} count={10} />
    </div>
  )
}
```

## API Usage

```bash
# Preview schedule
curl -X POST http://localhost:8080/api/v1/schedules/preview \
  -H "Content-Type: application/json" \
  -d '{
    "cron_expression": "0 9 * * 1-5",
    "timezone": "America/New_York",
    "count": 10
  }'
```

## Files Created/Modified

### Backend
- ✏️ `internal/schedule/service.go` - Added GetNextRunTimes method
- ✏️ `internal/schedule/service_test.go` - Added comprehensive tests
- ✏️ `internal/api/handlers/schedule.go` - Added PreviewSchedule handler
- ✏️ `internal/api/app.go` - Added /preview route
- ✏️ `internal/schedule/bulk_service.go` - Fixed compilation issues

### Frontend
- ✏️ `web/src/api/schedules.ts` - Added preview method
- ✏️ `web/src/api/schedules.test.ts` - Added preview tests
- ✨ `web/src/components/schedule/CronBuilder.tsx` - New component
- ✨ `web/src/components/schedule/CronBuilder.test.tsx` - New tests
- ✨ `web/src/components/schedule/SchedulePreview.tsx` - New component
- ✨ `web/src/components/schedule/SchedulePreview.test.tsx` - New tests
- ✨ `web/src/components/schedule/TimezoneSelector.tsx` - New component
- ✨ `web/src/components/schedule/TimezoneSelector.test.tsx` - New tests
- ✨ `web/src/components/schedule/ScheduleBuilder.example.tsx` - Example usage
- ✨ `web/src/components/schedule/index.ts` - Component exports

## Next Steps

To integrate into existing schedule creation flow:

1. Import components in schedule creation page:
   ```typescript
   import { CronBuilder, TimezoneSelector, SchedulePreview } from '@/components/schedule'
   ```

2. Replace manual cron input with visual builder

3. Add schedule name and enabled fields

4. Call `scheduleAPI.create()` on submit

## Accessibility

- ✅ All inputs have proper labels
- ✅ Keyboard navigation supported
- ✅ ARIA attributes for screen readers
- ✅ Error messages clearly communicated
- ✅ Loading states announced

## Browser Compatibility

- Modern browsers (Chrome, Firefox, Safari, Edge)
- Uses `date-fns` for reliable date formatting
- No IE11 support (uses modern ES6+ features)

## Performance

- ✅ Debouncing on search input (TimezoneSelector)
- ✅ Memoized calculations (CronBuilder description)
- ✅ Optimized re-renders with proper dependency arrays
- ✅ Lightweight components (< 300 lines each)
