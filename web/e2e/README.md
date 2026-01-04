# Gorax E2E Tests

Comprehensive end-to-end tests for the Gorax frontend using Playwright.

## Test Coverage

- ✅ **Workflows**: Create, edit, delete, execute workflows
- ✅ **Webhooks**: Create, view, test webhooks
- ✅ **Marketplace**: Browse, install, rate templates
- ✅ **Credentials**: Create, edit, delete credentials
- ✅ **Schedules**: Create, edit, delete, enable/disable schedules
- ✅ **Executions**: View, filter, retry executions
- ✅ **Analytics**: View dashboard, charts, metrics
- ✅ **OAuth**: Connect/disconnect OAuth providers
- ✅ **SSO**: Configure SSO providers
- ✅ **Audit Logs**: View, filter, export audit logs
- ✅ **Critical Flows**: Complete user journeys

## Prerequisites

1. Install dependencies:
```bash
cd web
npm install
```

2. Install Playwright browsers:
```bash
npx playwright install chromium
```

3. Ensure backend is running on `http://localhost:8080`

## Running Tests

### Run all tests (headless)
```bash
npm run test:e2e
```

### Run tests with UI (interactive)
```bash
npm run test:e2e:ui
```

### Run tests in headed mode (see browser)
```bash
npm run test:e2e:headed
```

### Run specific test file
```bash
npx playwright test workflows.spec.ts
```

### Run tests in debug mode
```bash
npm run test:e2e:debug
```

### View test report
```bash
npm run test:e2e:report
```

## Test Structure

```
tests/e2e/
├── setup.ts              # Test fixtures and helpers
├── utils/                # Utility functions
│   └── test-helpers.ts   # Common test utilities
├── workflows.spec.ts     # Workflow tests
├── webhooks.spec.ts      # Webhook tests
├── marketplace.spec.ts   # Marketplace tests
├── credentials.spec.ts   # Credentials tests
├── schedules.spec.ts     # Schedules tests
├── executions.spec.ts    # Executions tests
├── analytics.spec.ts     # Analytics tests
├── oauth.spec.ts         # OAuth tests
├── sso.spec.ts           # SSO tests
├── audit.spec.ts         # Audit logs tests
└── critical-flows.spec.ts # Critical user flows
```

## Configuration

Test configuration is in `playwright.config.ts`:
- Base URL: `http://localhost:5173`
- Timeout: 60 seconds per test
- Retries: 2 on CI, 0 locally
- Screenshots: On failure
- Video: On failure
- Trace: On failure

## Environment Variables

```bash
# Override base URL
PLAYWRIGHT_BASE_URL=http://localhost:3000

# Override API URL
API_BASE_URL=http://localhost:8080
```

## Authentication

Tests use fixtures for authentication:
- `authenticatedPage`: Regular user session
- `adminPage`: Admin user session

## Best Practices

1. **Use data-testid attributes** for stable selectors
2. **Wait for loading states** before assertions
3. **Handle empty states** gracefully
4. **Take screenshots on failure** for debugging
5. **Test real user flows** end-to-end

## Troubleshooting

### Tests failing with "Timeout"
- Increase timeout in `playwright.config.ts`
- Check if backend is running
- Check network connectivity

### Tests failing with "Element not found"
- Add `data-testid` attributes to components
- Use `waitForLoading()` helper
- Check if feature is behind feature flag

### Screenshots not captured
- Check `playwright-report/` directory
- Run with `--trace on` for full traces

## CI Integration

Tests run automatically in CI with:
- 2 retries on failure
- Full traces on failure
- HTML report artifact

## Writing New Tests

1. Follow existing test patterns
2. Use helpers from `setup.ts`
3. Add `data-testid` to new components
4. Test happy path AND error cases
5. Clean up test data after tests
