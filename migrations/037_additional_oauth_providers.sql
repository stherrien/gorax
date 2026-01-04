-- migrations/037_additional_oauth_providers.sql
-- Add Twitter, LinkedIn, Salesforce, and Auth0 OAuth providers

-- Insert new OAuth providers with proper configuration
INSERT INTO oauth_providers (provider_key, name, description, auth_url, token_url, user_info_url, default_scopes, status) VALUES
    (
        'twitter',
        'Twitter',
        'Twitter/X API integration with OAuth 2.0',
        'https://api.twitter.com/2/oauth2/authorize',
        'https://api.twitter.com/2/oauth2/token',
        'https://api.twitter.com/2/users/me',
        ARRAY['tweet.read', 'users.read', 'offline.access'],
        'active'
    ),
    (
        'linkedin',
        'LinkedIn',
        'LinkedIn professional network integration',
        'https://www.linkedin.com/oauth/v2/authorization',
        'https://www.linkedin.com/oauth/v2/accessToken',
        'https://api.linkedin.com/v2/userinfo',
        ARRAY['profile', 'email', 'openid'],
        'active'
    ),
    (
        'salesforce',
        'Salesforce',
        'Salesforce CRM and platform integration',
        'https://login.salesforce.com/services/oauth2/authorize',
        'https://login.salesforce.com/services/oauth2/token',
        'https://login.salesforce.com/services/oauth2/userinfo',
        ARRAY['api', 'refresh_token', 'openid', 'profile', 'email'],
        'active'
    ),
    (
        'auth0',
        'Auth0',
        'Auth0 identity platform integration (requires tenant-specific domain configuration)',
        '{custom}',
        '{custom}',
        '{custom}',
        ARRAY['openid', 'profile', 'email', 'offline_access'],
        'active'
    )
ON CONFLICT (provider_key) DO NOTHING;

-- Comments for documentation
COMMENT ON TABLE oauth_providers IS 'OAuth 2.0 provider configurations - now includes Twitter, LinkedIn, Salesforce, and Auth0';

-- Provider-specific notes
COMMENT ON COLUMN oauth_providers.auth_url IS 'Authorization endpoint URL. Use {custom} for tenant-specific providers like Auth0';
COMMENT ON COLUMN oauth_providers.config IS 'Provider-specific configuration (e.g., Salesforce sandbox vs production, Auth0 tenant domain)';
