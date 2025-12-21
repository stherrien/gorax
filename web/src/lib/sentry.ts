import * as Sentry from '@sentry/react';

// Sentry configuration interface
export interface SentryConfig {
  dsn: string;
  environment: string;
  sampleRate?: number;
  tracesSampleRate?: number;
  enabled?: boolean;
}

// Initialize Sentry error tracking
export function initializeSentry(config: SentryConfig): void {
  // Don't initialize if disabled or no DSN provided
  if (!config.enabled || !config.dsn) {
    console.info('Sentry error tracking is disabled');
    return;
  }

  Sentry.init({
    dsn: config.dsn,
    environment: config.environment,
    integrations: [
      Sentry.browserTracingIntegration(),
      Sentry.replayIntegration({
        // Mask all text content, but still capture DOM structure
        maskAllText: true,
        blockAllMedia: true,
      }),
    ],

    // Performance monitoring
    tracesSampleRate: config.tracesSampleRate ?? 1.0,

    // Error sampling
    sampleRate: config.sampleRate ?? 1.0,

    // Session Replay (captures 10% of sessions, 100% of error sessions)
    replaysSessionSampleRate: 0.1,
    replaysOnErrorSampleRate: 1.0,

    // Configure which origins are allowed for error tracking
    allowUrls: [
      window.location.origin,
      /^https:\/\/.*\.gorax\.io/,
    ],

    // Ignore certain errors
    ignoreErrors: [
      // Browser extensions
      'top.GLOBALS',
      // Random plugins/extensions
      'originalCreateNotification',
      'canvas.contentDocument',
      'MyApp_RemoveAllHighlights',
      // Network errors
      'Network request failed',
      'NetworkError',
      // Aborted requests
      'Request aborted',
      'AbortError',
    ],

    // Filter out known third-party script errors
    beforeSend(event, _hint) {
      // Filter out errors from browser extensions
      if (event.exception?.values?.[0]?.stacktrace?.frames) {
        const frames = event.exception.values[0].stacktrace.frames;
        const isBrowserExtension = frames.some((frame) => {
          const filename = frame.filename || '';
          return filename.includes('extension://') ||
                 filename.includes('chrome-extension://') ||
                 filename.includes('moz-extension://');
        });

        if (isBrowserExtension) {
          return null; // Don't send
        }
      }

      return event;
    },
  });
}

// Set user context for error tracking
export function setUser(user: {
  id: string;
  email?: string;
  username?: string;
}): void {
  Sentry.setUser({
    id: user.id,
    email: user.email,
    username: user.username,
  });
}

// Clear user context (on logout)
export function clearUser(): void {
  Sentry.setUser(null);
}

// Add custom context/tags
export function setContext(key: string, value: Record<string, any>): void {
  Sentry.setContext(key, value);
}

export function setTag(key: string, value: string): void {
  Sentry.setTag(key, value);
}

// Manually capture an error
export function captureError(error: Error, context?: Record<string, any>): void {
  Sentry.captureException(error, {
    contexts: context ? { custom: context } : undefined,
  });
}

// Manually capture a message
export function captureMessage(
  message: string,
  level: 'fatal' | 'error' | 'warning' | 'log' | 'info' | 'debug' = 'info'
): void {
  Sentry.captureMessage(message, level);
}

// Add breadcrumb (audit trail of events)
export function addBreadcrumb(breadcrumb: {
  message: string;
  category?: string;
  level?: 'fatal' | 'error' | 'warning' | 'log' | 'info' | 'debug';
  data?: Record<string, any>;
}): void {
  Sentry.addBreadcrumb({
    message: breadcrumb.message,
    category: breadcrumb.category || 'custom',
    level: breadcrumb.level || 'info',
    data: breadcrumb.data,
  });
}

// Create a scoped error tracker for workflow executions
export function withExecutionScope<T>(
  executionId: string,
  workflowId: string,
  fn: () => T
): T {
  return Sentry.withScope((scope) => {
    scope.setTag('execution_id', executionId);
    scope.setTag('workflow_id', workflowId);
    scope.setContext('execution', {
      executionId,
      workflowId,
    });
    return fn();
  });
}

// Create a scoped error tracker for tenant operations
export function withTenantScope<T>(
  tenantId: string,
  fn: () => T
): T {
  return Sentry.withScope((scope) => {
    scope.setTag('tenant_id', tenantId);
    scope.setContext('tenant', { tenantId });
    return fn();
  });
}

// Export Sentry for advanced use cases
export { Sentry };
