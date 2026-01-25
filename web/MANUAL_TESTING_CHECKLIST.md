# Manual Testing Checklist for Routing Fixes

## Pre-Testing Setup

1. Start the development server:
   ```bash
   npm run dev
   ```

2. Open browser console (F12)
3. Open network tab
4. Clear browser cache

## Critical Test Cases

### ❌ BEFORE FIX (What Was Broken)
- Navigating to `/webhooks/new` would make API call to `/api/v1/webhooks/new`
- Console would show 404 errors
- Page would show "Webhook not found"

### ✅ AFTER FIX (Expected Behavior)
- Navigating to `/webhooks/new` should render create form
- No API calls with `id="new"` in the URL
- No console errors
- Page renders immediately

## Test Routes - All Must Pass

### 1. Webhooks
- [ ] Navigate to `/webhooks` → Should show webhook list
- [ ] Navigate to `/webhooks/new` → Should NOT make API call to `/api/v1/webhooks/new`
- [ ] Navigate to `/webhooks/550e8400-e29b-41d4-a716-446655440000` → Should fetch valid webhook
- [ ] Navigate to `/webhooks/invalid` → Should redirect to `/webhooks`
- [ ] Navigate to `/webhooks/create` → Should redirect to `/webhooks`

**Console Check**: Zero errors
**Network Tab**: Only valid API calls (no `/api/v1/webhooks/new` or `/api/v1/webhooks/create`)

### 2. Workflows
- [ ] Navigate to `/workflows` → Should show workflow list
- [ ] Navigate to `/workflows/new` → Should show empty editor (no API call)
- [ ] Navigate to `/workflows/550e8400-e29b-41d4-a716-446655440000` → Should load workflow
- [ ] Navigate to `/workflows/invalid` → Should show error state (no API call)
- [ ] Navigate to `/workflows/edit` → Should show error state (no API call)

**Console Check**: Zero errors
**Network Tab**: No API calls to `/api/v1/workflows/new` or `/api/v1/workflows/edit`

### 3. Executions
- [ ] Navigate to `/executions` → Should show execution list
- [ ] Navigate to `/executions/550e8400-e29b-41d4-a716-446655440000` → Should load execution
- [ ] Navigate to `/executions/new` → Should redirect to `/executions`
- [ ] Navigate to `/executions/invalid` → Should redirect to `/executions`

**Console Check**: Zero errors
**Network Tab**: No API calls to `/api/v1/executions/new`

### 4. Schedules
- [ ] Navigate to `/schedules` → Should show schedule list
- [ ] Navigate to `/schedules/new` → Should show create form
- [ ] Navigate to `/schedules/550e8400-e29b-41d4-a716-446655440000/edit` → Should load schedule
- [ ] Navigate to `/schedules/new/edit` → Should redirect to `/schedules`
- [ ] Navigate to `/schedules/invalid/edit` → Should redirect to `/schedules`

**Console Check**: Zero errors
**Network Tab**: No API calls to `/api/v1/schedules/new` or `/api/v1/schedules/invalid`

### 5. Other Routes (Baseline Check)
- [ ] Navigate to `/` → Dashboard loads
- [ ] Navigate to `/credentials` → Credential manager loads
- [ ] Navigate to `/marketplace` → Marketplace loads
- [ ] Navigate to `/analytics` → Analytics dashboard loads
- [ ] Navigate to `/oauth/connections` → OAuth connections load
- [ ] Navigate to `/admin/sso` → SSO settings load
- [ ] Navigate to `/admin/audit-logs` → Audit logs load
- [ ] Navigate to `/docs` → Documentation loads
- [ ] Navigate to `/ai/builder` → AI workflow builder loads

**Console Check**: Zero errors for all pages

## Browser Navigation Tests

### Back/Forward Button
- [ ] Navigate: `/workflows` → `/workflows/new` → Browser back button
  - Expected: Returns to `/workflows` without errors
- [ ] Navigate: `/webhooks` → `/webhooks/new` → `/webhooks/:id` → Browser back button twice
  - Expected: Returns through history without errors

### Direct URL Entry
- [ ] Type `/webhooks/new` directly in address bar → Press Enter
  - Expected: No 404, no API call to `/api/v1/webhooks/new`
- [ ] Type `/workflows/create` directly in address bar → Press Enter
  - Expected: Redirects or shows error (no API call)

### Page Refresh
- [ ] Navigate to `/workflows/new` → Press F5 to refresh
  - Expected: Page reloads without errors, no API calls

## Performance Tests

### Page Load Times
- [ ] All routes load in < 2 seconds
- [ ] No unnecessary API calls during page load
- [ ] No visible loading flicker

### Network Efficiency
- [ ] Each detail page makes exactly 1 API call (for valid IDs)
- [ ] List pages don't make individual item API calls
- [ ] No duplicate or retry API calls

## Edge Cases

### Special Characters in URL
- [ ] Navigate to `/webhooks/test-123` → Should redirect/error (no API call)
- [ ] Navigate to `/workflows/%20` → Should redirect/error (no API call)
- [ ] Navigate to `/executions/null` → Should redirect/error (no API call)

### Case Sensitivity
- [ ] Navigate to `/webhooks/NEW` → Should redirect/error (reserved keyword)
- [ ] Navigate to `/webhooks/New` → Should redirect/error (reserved keyword)
- [ ] Navigate to `/workflows/EDIT` → Should redirect/error (reserved keyword)

## Console Verification

Open browser console and check for:
- [ ] ❌ ZERO errors
- [ ] ❌ ZERO warnings about invalid IDs
- [ ] ❌ ZERO 404 responses
- [ ] ❌ ZERO failed API calls

## Network Tab Verification

Check Network tab (filter by "Fetch/XHR"):
- [ ] ✅ Only valid UUID IDs in API URLs
- [ ] ❌ No `/api/v1/webhooks/new`
- [ ] ❌ No `/api/v1/webhooks/create`
- [ ] ❌ No `/api/v1/workflows/new`
- [ ] ❌ No `/api/v1/executions/new`
- [ ] ❌ No `/api/v1/schedules/new`
- [ ] ❌ No API calls with invalid UUIDs

## Regression Tests

### Existing Functionality Still Works
- [ ] Can create new workflow via `/workflows/new`
- [ ] Can edit existing workflow via `/workflows/:id`
- [ ] Can view webhook details via `/webhooks/:id`
- [ ] Can create new schedule via `/schedules/new`
- [ ] Can edit schedule via `/schedules/:id/edit`

### User Workflows
- [ ] Create Workflow: Navigate to `/workflows/new` → Add nodes → Save → Redirects to `/workflows/:id`
- [ ] Edit Workflow: Navigate to `/workflows/:id` → Modify → Save → Success message
- [ ] View Execution: Click execution from list → Details page loads
- [ ] Create Schedule: Navigate to `/schedules/new` → Fill form → Submit → Success

## Sign-Off

### Developer Testing
- [ ] All critical test cases pass
- [ ] All edge cases handled
- [ ] No console errors
- [ ] No invalid API calls
- [ ] Performance is acceptable

### QA Testing
- [ ] Full regression test complete
- [ ] All user workflows tested
- [ ] Cross-browser testing (Chrome, Firefox, Safari)
- [ ] Mobile responsive testing

### Ready for Production
- [ ] All tests pass
- [ ] No blocking issues found
- [ ] Performance benchmarks met
- [ ] Documentation updated

## Troubleshooting

### If You See Errors:

**404 Not Found for /api/v1/webhooks/new**
- ❌ FAIL: Route ordering is still broken
- Fix: Check `/web/src/App.tsx` route order

**Console Error: "Invalid ID: new"**
- ❌ FAIL: ID validation not working
- Fix: Check component is using `isValidResourceId()`

**Page Shows "Not Found" for /workflows/new**
- ❌ FAIL: Route definition missing
- Fix: Verify route exists in App.tsx

**API Calls Still Happening with id="new"**
- ❌ FAIL: Hook validation not applied
- Fix: Check hook has `enabled: isValidResourceId(id)`

## Success Criteria

All checkboxes above must be checked (✅) with:
- Zero console errors
- Zero invalid API calls
- All pages load correctly
- All user workflows function

**ONLY THEN** is this ready for production deployment.
