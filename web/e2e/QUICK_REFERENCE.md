# E2E Tests Quick Reference

## Run Tests

```bash
# All tests
npm run test:e2e

# With UI (best for development)
npm run test:e2e:ui

# See browser
npm run test:e2e:headed

# Specific file
npx playwright test workflows.spec.ts

# Specific test
npx playwright test -g "should create"

# Debug
npm run test:e2e:debug
```

## View Results

```bash
# HTML report
npm run test:e2e:report

# Screenshot
open test-results/.../test-failed-1.png

# Trace
npx playwright show-trace test-results/.../trace.zip
```

## Common data-testid Patterns

```tsx
// Cards
<div data-testid="workflow-card">
<div data-testid="webhook-card">
<div data-testid="credential-card">

// Inputs
<input data-testid="workflow-name-input" name="name" />
<textarea data-testid="description-input" name="description" />

// Buttons
<button data-testid="save-button">Save</button>
<button data-testid="delete-button">Delete</button>
<button data-testid="create-button">Create</button>

// Editor
<div data-testid="workflow-editor">

// Toasts
<div data-testid="toast-success">
<div data-testid="toast-error">

// Tables
<table data-testid="executions-table">
<tr data-testid="execution-row">

// Empty states
<div data-testid="empty-state">
```

## Test Structure

```typescript
test('should do something', async ({ authenticatedPage: page }) => {
  // Arrange
  await navigateTo(page, 'Workflows')
  
  // Act
  await clickButton(page, 'Create')
  await fillFormField(page, 'Name', 'Test')
  await clickButton(page, 'Save')
  
  // Assert
  await expectToast(page, /created/i)
})
```

## Helpers Available

```typescript
// Navigation
navigateTo(page, 'Workflows')
navigateAndWait(page, '/workflows')

// Forms
fillFormField(page, 'Name', 'value')
selectDropdownOption(page, 'Status', 'active')
clickButton(page, 'Save')

// Assertions
expectToast(page, /success/i)
expectError(page, /invalid/i)
expectTableRow(page, 'Test Workflow')
expectEmptyState(page)
expectVisible(page, '[data-testid="card"]')

// Utilities
waitForLoading(page)
waitForAPICall(page, '/api/workflows')
searchFor(page, 'test query')
```

## Fixtures

```typescript
test('as regular user', async ({ authenticatedPage }) => {
  // authenticatedPage has regular user logged in
})

test('as admin', async ({ adminPage }) => {
  // adminPage has admin user logged in
})
```

## Debugging Tips

1. **Use UI mode:** `npm run test:e2e:ui`
2. **Check screenshot:** Auto-captured on failure
3. **View trace:** Full recording of test
4. **Add console.log:** Shows in test output
5. **Use page.pause():** Stops test for inspection

## Common Issues

### "Element not found"
→ Add `data-testid` attribute to component

### "Timeout"
→ Check if backend is running
→ Increase timeout in helper

### "Strict mode violation"
→ Selector matches multiple elements
→ Make selector more specific

### Tests pass locally but fail in CI
→ Add retry logic
→ Increase timeouts
→ Check for race conditions

## File Locations

- Tests: `/web/e2e/*.spec.ts`
- Config: `/web/playwright.config.ts`
- Helpers: `/web/e2e/utils/test-helpers.ts`
- Fixtures: `/web/e2e/setup.ts`

## Coverage

✅ Workflows (10 tests)
✅ Webhooks (9 tests)
✅ Marketplace (12 tests)
✅ Credentials (3 tests)
✅ Schedules (6 tests)
✅ Executions (8 tests)
✅ Analytics (7 tests)
✅ OAuth (6 tests)
✅ SSO (6 tests)
✅ Audit Logs (16 tests)
✅ Critical Flows (2 tests)

**Total: 84 tests**
