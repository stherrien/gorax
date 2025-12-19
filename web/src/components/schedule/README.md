# Schedule Components Integration

This directory contains the schedule management components for the Gorax workflow automation platform.

## Components

### CronBuilder
Visual cron expression builder with presets and advanced mode.
- Simple mode: Grid-based input for cron parts
- Advanced mode: Direct cron string input
- Presets: Common schedules (hourly, daily, weekly, etc.)
- Validation and human-readable descriptions

### ScheduleForm
Complete form component for creating/editing schedules. Integrates:
- CronBuilder for cron expression input
- TimezoneSelector for timezone selection
- SchedulePreview for showing next run times
- Form validation and error handling

**Usage:**
```tsx
import { ScheduleForm } from './components/schedule'

function MyComponent() {
  const handleSubmit = async (data) => {
    await scheduleAPI.create(workflowId, data)
  }

  return (
    <ScheduleForm
      onSubmit={handleSubmit}
      onCancel={() => navigate('/schedules')}
    />
  )
}
```

### TimezoneSelector
Dropdown selector for timezones with search and current time display.

### SchedulePreview
Shows next N execution times for a cron expression in a given timezone.

## Pages

### CreateSchedule
Page for creating new schedules. Requires `workflowId` as a URL parameter.

**Route:** `/schedules/new?workflowId=<id>`

## Testing

All components have comprehensive test coverage following TDD principles:
- ScheduleForm.test.tsx: 19 tests covering form interaction, validation, and submission
- CreateSchedule.test.tsx: 5 tests covering page render and schedule creation flow

Run tests:
```bash
npm test -- src/components/schedule
npm test -- src/pages/CreateSchedule.test.tsx
```

## Integration Points

1. **Schedules Page** (`/schedules`) - Lists all schedules with link to create new ones
2. **Create Schedule Page** (`/schedules/new`) - Form for creating new schedules
3. **API** (`api/schedules.ts`) - REST API client for schedule operations
4. **Hooks** (`hooks/useSchedules.ts`) - React hooks for schedule data management
