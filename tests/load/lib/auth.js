// Authentication Helper for k6 Load Tests
// Supports both DevAuth (development) and Kratos (production) authentication modes

import http from 'k6/http';

/**
 * Setup authentication based on environment configuration
 *
 * Returns authentication data object containing:
 * - mode: 'dev' or 'kratos'
 * - For dev mode: tenantID and userID
 * - For kratos mode: sessionToken (to be implemented)
 *
 * Environment Variables:
 * - AUTH_MODE: 'dev' or 'kratos' (default: 'dev')
 * - TEST_USER_ID: User ID for DevAuth (default: 'default-test-user')
 * - TEST_TENANT_ID: Tenant ID for DevAuth (default: 'default-tenant')
 * - KRATOS_PUBLIC_URL: Kratos public URL for production (default: 'http://localhost:4433')
 *
 * @param {Object} config - Configuration object from config.js
 * @returns {Object} Authentication data for use in test functions
 * @throws {Error} If authentication setup fails
 */
export function setupAuth(config) {
  const mode = config?.auth?.mode || __ENV.AUTH_MODE || 'dev';

  if (mode === 'dev') {
    console.log('✓ Using DevAuth mode (X-Tenant-ID and X-User-ID headers)');

    const tenantID = config?.auth?.devTenantID || __ENV.TEST_TENANT_ID || 'default-tenant';
    const userID = config?.auth?.devUserID || __ENV.TEST_USER_ID || 'default-test-user';

    console.log(`  Tenant ID: ${tenantID}`);
    console.log(`  User ID: ${userID}`);

    return {
      mode: 'dev',
      tenantID,
      userID,
    };
  } else if (mode === 'kratos') {
    console.log('✓ Using Kratos authentication mode');

    // Kratos authentication flow
    const sessionToken = authenticateWithKratos(config);

    return {
      mode: 'kratos',
      sessionToken,
    };
  }

  throw new Error(`Unknown auth mode: ${mode}. Supported modes: 'dev', 'kratos'`);
}

/**
 * Get HTTP headers for authenticated requests
 *
 * @param {Object} authData - Authentication data from setupAuth()
 * @returns {Object} HTTP headers object
 */
export function getAuthHeaders(authData) {
  if (!authData) {
    throw new Error('authData is required. Did you call setupAuth() in the setup function?');
  }

  if (authData.mode === 'dev') {
    return {
      'X-Tenant-ID': authData.tenantID,
      'X-User-ID': authData.userID,
    };
  } else if (authData.mode === 'kratos') {
    return {
      'Authorization': `Bearer ${authData.sessionToken}`,
    };
  }

  return {};
}

/**
 * Get headers for DevAuth mode (direct access, useful for simple tests)
 *
 * @param {string} tenantID - Tenant ID (default: 'default-tenant')
 * @param {string} userID - User ID (default: 'default-test-user')
 * @returns {Object} HTTP headers object with X-Tenant-ID and X-User-ID
 */
export function getDevAuthHeaders(tenantID, userID) {
  return {
    'X-Tenant-ID': tenantID || __ENV.TEST_TENANT_ID || 'default-tenant',
    'X-User-ID': userID || __ENV.TEST_USER_ID || 'default-test-user',
  };
}

/**
 * Get headers for Kratos authentication (placeholder for future implementation)
 *
 * @param {string} sessionToken - Kratos session token
 * @returns {Object} HTTP headers object with Authorization bearer token
 */
export function getKratosAuthHeaders(sessionToken) {
  if (!sessionToken) {
    throw new Error('sessionToken is required for Kratos authentication');
  }

  return {
    'Authorization': `Bearer ${sessionToken}`,
  };
}

/**
 * Authenticate with Ory Kratos (production authentication)
 *
 * This function implements the Kratos authentication flow:
 * 1. Initialize login flow
 * 2. Submit credentials
 * 3. Extract session token
 *
 * @param {Object} config - Configuration object from config.js
 * @returns {string} Session token
 * @throws {Error} If authentication fails
 *
 * @todo Implement full Kratos authentication flow when needed for production testing
 */
function authenticateWithKratos(config) {
  const kratosURL = config?.auth?.kratosPublicURL || __ENV.KRATOS_PUBLIC_URL || 'http://localhost:4433';

  console.log(`  Kratos URL: ${kratosURL}`);
  console.log('  ⚠ Kratos authentication not yet implemented');

  // TODO: Implement Kratos authentication flow
  // Step 1: Initialize login flow
  // const initRes = http.get(`${kratosURL}/self-service/login/browser`);
  //
  // Step 2: Extract flow ID and CSRF token
  // const flowID = initRes.json('id');
  // const csrfToken = initRes.json('ui.nodes[?(@.attributes.name=="csrf_token")].attributes.value');
  //
  // Step 3: Submit credentials
  // const loginRes = http.post(`${kratosURL}/self-service/login?flow=${flowID}`, {
  //   method: 'password',
  //   csrf_token: csrfToken,
  //   identifier: config.testUser.email,
  //   password: config.testUser.password,
  // });
  //
  // Step 4: Extract session token from cookie or response
  // const sessionToken = loginRes.cookies['ory_kratos_session'];
  //
  // return sessionToken;

  throw new Error('Kratos authentication not yet implemented. Use AUTH_MODE=dev for development testing.');
}

/**
 * Verify that authentication is working by making a test request
 *
 * @param {string} baseUrl - Base URL of the API
 * @param {Object} authData - Authentication data from setupAuth()
 * @returns {boolean} True if authentication is working
 */
export function verifyAuth(baseUrl, authData) {
  const headers = {
    'Content-Type': 'application/json',
    ...getAuthHeaders(authData),
  };

  const res = http.get(`${baseUrl}/api/v1/workflows?limit=1`, { headers });

  if (res.status === 200) {
    console.log('✓ Authentication verified successfully');
    return true;
  } else {
    console.error(`✗ Authentication verification failed: ${res.status} - ${res.body}`);
    return false;
  }
}

export default {
  setupAuth,
  getAuthHeaders,
  getDevAuthHeaders,
  getKratosAuthHeaders,
  verifyAuth,
};
