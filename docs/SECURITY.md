# Security Architecture and Guidelines

## Overview

Gorax is a workflow automation platform that handles sensitive data including user credentials, webhook secrets, and workflow execution data. This document outlines the security architecture, authentication flows, and best practices for secure deployment.

## Security Principles

1. **Defense in Depth**: Multiple layers of security controls
2. **Least Privilege**: Minimal permissions by default
3. **Zero Trust**: Verify all requests regardless of source
4. **Encryption Everywhere**: Data encrypted at rest and in transit
5. **Audit Logging**: Comprehensive logging of security events

## Architecture Overview

### Components

```
┌─────────────────┐
│   Web Frontend  │ (React/TypeScript)
└────────┬────────┘
         │ HTTPS
         ▼
┌─────────────────┐
│   API Server    │ (Go/Chi Router)
│  - Authentication│
│  - Authorization│
│  - Rate Limiting│
└────────┬────────┘
         │
    ┌────┼────────────────┐
    │    │                │
    ▼    ▼                ▼
┌────────┐  ┌──────────┐  ┌──────────┐
│Postgres│  │  Redis   │  │ Ory      │
│        │  │(Sessions)│  │ Kratos   │
└────────┘  └──────────┘  └──────────┘
```

## Authentication and Authorization

### Authentication Flow

Gorax uses **Ory Kratos** for authentication, providing:

- **Password-based authentication** with secure hashing (Argon2id)
- **Email verification** for new accounts
- **Password reset** via secure tokens
- **Session management** with HTTP-only cookies
- **OAuth 2.0/OIDC** support for Google, GitHub, Microsoft

#### Authentication Process

1. **User Registration**:
   ```
   Client → API /auth/register/initiate → Kratos (create flow)
   Client → API /auth/register → Kratos (submit credentials)
   Kratos → Webhook → API (sync user to database)
   ```

2. **User Login**:
   ```
   Client → API /auth/login/initiate → Kratos (create flow)
   Client → API /auth/login → Kratos (verify credentials)
   Kratos → Sets session cookie (ory_kratos_session)
   ```

3. **Session Validation**:
   ```
   Client → API (with session cookie)
   API Middleware → Kratos /sessions/whoami
   Kratos → Returns user identity + traits
   API → Extracts tenant_id → Continues request
   ```

### Authorization Model

#### Multi-Tenancy

- **Tenant Isolation**: All data is scoped to tenant_id
- **Row-Level Security**: Database queries enforce tenant boundaries
- **Context Propagation**: Tenant ID extracted from user session

#### Role-Based Access Control (RBAC)

Roles available in the system:

| Role | Permissions | Description |
|------|-------------|-------------|
| **admin** | All operations, cross-tenant access | System administrators |
| **owner** | All operations within tenant | Tenant owner |
| **editor** | Create, read, update workflows | Standard user |
| **viewer** | Read-only access | Auditor/observer |

#### Permission Matrix

| Resource | Admin | Owner | Editor | Viewer |
|----------|-------|-------|--------|--------|
| Create Workflows | ✓ | ✓ | ✓ | ✗ |
| Read Workflows | ✓ | ✓ | ✓ | ✓ |
| Update Workflows | ✓ | ✓ | ✓ | ✗ |
| Delete Workflows | ✓ | ✓ | ✗ | ✗ |
| Execute Workflows | ✓ | ✓ | ✓ | ✗ |
| Manage Credentials | ✓ | ✓ | ✓ | ✗ |
| View Credentials | ✓ | ✓ | ✓ | ✗ |
| Access Credential Values | ✓ | ✓ | ✓ | ✗ |
| Manage Webhooks | ✓ | ✓ | ✓ | ✗ |
| View Executions | ✓ | ✓ | ✓ | ✓ |
| Manage Tenants | ✓ | ✗ | ✗ | ✗ |

### Development Mode Authentication

For development convenience, a simplified auth mode is available:

```go
// internal/api/middleware/dev_auth.go
// Sets default user with dev-tenant-1 without requiring Kratos
```

**⚠️ WARNING**: NEVER use development mode in production. Set `APP_ENV=production` and configure Kratos properly.

## Credential Management

### Encryption Architecture

Gorax implements **envelope encryption** for credential storage:

```
┌─────────────────────────────────────────────┐
│          Envelope Encryption Flow           │
└─────────────────────────────────────────────┘

Plaintext Credential
        ↓
    Generate DEK (Data Encryption Key)
        ↓
    Encrypt credential with DEK (AES-256-GCM)
        ↓
    Encrypt DEK with Master Key (AES-256-GCM)
        ↓
    Store: Encrypted DEK + Encrypted Credential + Nonce + Auth Tag
```

### Encryption Details

**Algorithm**: AES-256-GCM (Galois/Counter Mode)
- **Key Size**: 256 bits (32 bytes)
- **Nonce**: 96 bits (12 bytes), cryptographically random
- **Authentication Tag**: 128 bits (16 bytes)
- **Benefits**: Provides both confidentiality and authenticity

**Key Management**:

1. **Development**: Master key stored in environment variable (base64-encoded)
2. **Production** (Recommended): AWS KMS for master key management
   - Key rotation support
   - Audit trail via CloudTrail
   - Hardware Security Module (HSM) backed

### Credential Access Logging

All credential access is logged with:
- User ID and tenant ID
- Timestamp
- Operation type (read, create, update, delete, rotate)
- IP address
- User agent

### Credential Masking

Sensitive values are masked in logs and API responses:

```go
// Example: API key masking
"sk_xxxx_EXAMPLE_KEY_REDACTED" → "sk_xxxx_***************CTED"
```

Masking rules:
- **API Keys**: Show first 8 and last 4 characters
- **Tokens**: Show first 6 and last 4 characters
- **Passwords**: Never logged or displayed
- **Secrets**: Show first 4 and last 4 characters

## Data Encryption

### Encryption at Rest

| Data Type | Encryption Method | Key Management |
|-----------|------------------|----------------|
| Credentials | AES-256-GCM | Master key (KMS in prod) |
| Database | PostgreSQL encryption | Database-level encryption |
| Redis Cache | Redis AUTH + TLS | Environment variable |
| S3 Artifacts | S3 Server-Side Encryption | AWS-managed or KMS |

### Encryption in Transit

All network communication must use TLS 1.2 or higher:

- **API → Client**: HTTPS (TLS 1.2+)
- **API → Database**: PostgreSQL TLS
- **API → Redis**: Redis TLS (optional, recommended)
- **API → Kratos**: HTTPS
- **Worker → SQS**: AWS Signature V4 over HTTPS

## Webhook Security

### Webhook Signature Verification

Webhooks use HMAC-SHA256 for request verification:

```
HMAC-SHA256(webhook_secret, request_body) = signature
```

Signature is sent in the `X-Webhook-Signature` header:

```http
X-Webhook-Signature: sha256=<hex_encoded_signature>
```

**Verification Process**:

1. Extract signature from header
2. Compute HMAC-SHA256 of request body with webhook secret
3. Compare signatures using constant-time comparison
4. Reject request if signatures don't match

### Webhook Authentication Types

1. **No Authentication**: Accept all requests (development only)
2. **Secret-based**: HMAC signature verification
3. **Custom Headers**: Verify specific headers (API key, etc.)

### Webhook Event Storage

- Events stored for audit trail
- Retention period: 30 days (configurable)
- Includes request headers, body, and response status
- Automatic cleanup via scheduled job

## Input Validation and Sanitization

### API Request Validation

All API inputs are validated before processing:

```go
// Example validation
type CreateWorkflowRequest struct {
    Name        string `json:"name" validate:"required,min=1,max=100"`
    Description string `json:"description" validate:"max=500"`
    TriggerType string `json:"trigger_type" validate:"required,oneof=webhook schedule manual"`
}
```

### SQL Injection Prevention

**All database queries use parameterized statements**:

```go
// ✓ SAFE: Parameterized query
query := `SELECT * FROM workflows WHERE id = $1 AND tenant_id = $2`
err := db.GetContext(ctx, &workflow, query, workflowID, tenantID)

// ✗ UNSAFE: String concatenation (NEVER DO THIS)
// query := fmt.Sprintf("SELECT * FROM workflows WHERE id = '%s'", workflowID)
```

### Path Traversal Prevention

- File paths are validated against allowed directories
- No user-supplied paths used for file operations
- Artifact downloads use UUID-based identifiers only

### XSS Prevention

Frontend security measures:

- React's built-in XSS protection (automatic escaping)
- Content Security Policy (CSP) headers
- No `dangerouslySetInnerHTML` usage
- Sanitization of user-generated content

## Rate Limiting and Quotas

### Rate Limiting

Implemented at multiple levels:

1. **Global Rate Limit**: 1000 requests/minute per IP
2. **Per-Tenant Limit**: Configurable (default: 100 req/min)
3. **Endpoint-Specific Limits**:
   - Login: 5 attempts/5 minutes per IP
   - Webhook: 100 requests/minute per webhook

### Quota Management

Tenants have configurable quotas:

- **Workflows**: Max number of workflows
- **Executions**: Max executions per month
- **Storage**: Max artifact storage (GB)
- **Webhooks**: Max active webhooks

Quotas are enforced via middleware:

```go
// Middleware checks quota before allowing request
func (q *QuotaChecker) CheckQuotas() func(next http.Handler) http.Handler
```

## Session Management

### Session Security

- **HTTP-Only Cookies**: Not accessible via JavaScript
- **Secure Flag**: Only sent over HTTPS
- **SameSite=Lax**: CSRF protection
- **Session Timeout**: 24 hours (configurable)
- **Absolute Timeout**: 7 days (configurable)

### Session Storage

Sessions managed by Ory Kratos:

- Stored in Kratos database (encrypted)
- Redis used for session metadata caching
- Automatic cleanup of expired sessions

## Security Headers

Recommended HTTP security headers:

```
Strict-Transport-Security: max-age=31536000; includeSubDomains
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-XSS-Protection: 1; mode=block
Content-Security-Policy: default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'
Referrer-Policy: strict-origin-when-cross-origin
Permissions-Policy: geolocation=(), microphone=(), camera=()
```

## CORS Configuration

CORS is configured restrictively:

```go
cors.Options{
    AllowedOrigins:   []string{"https://app.gorax.io"},
    AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
    AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Tenant-ID"},
    ExposedHeaders:   []string{"Link"},
    AllowCredentials: true,
    MaxAge:           300,
}
```

**⚠️ Development**: Localhost origins are allowed in dev mode. Remove for production.

## Logging and Monitoring

### Security Event Logging

All security-relevant events are logged:

1. **Authentication Events**:
   - Login success/failure
   - Logout
   - Session expiration
   - Password changes

2. **Authorization Events**:
   - Permission denied
   - Role changes
   - Cross-tenant access attempts

3. **Credential Events**:
   - Credential creation
   - Credential access
   - Credential rotation
   - Credential deletion

4. **Webhook Events**:
   - Webhook creation
   - Signature verification failures
   - Suspicious request patterns

### Log Format

Structured logging using slog:

```json
{
  "timestamp": "2025-01-15T10:30:00Z",
  "level": "WARN",
  "msg": "credential access",
  "user_id": "usr_abc123",
  "tenant_id": "tenant_xyz",
  "credential_id": "cred_def456",
  "operation": "read",
  "ip": "192.168.1.1",
  "user_agent": "Mozilla/5.0..."
}
```

### Sensitive Data in Logs

**NEVER log**:
- Passwords or password hashes
- Session tokens
- API keys or secrets
- Credential values
- Webhook secrets

**Masked in logs**:
- User emails (partially masked)
- IP addresses (last octet masked in some contexts)

## Secure Deployment Checklist

### Pre-Production

- [ ] Change all default credentials
- [ ] Generate strong random secrets (use `openssl rand -base64 32`)
- [ ] Configure TLS certificates (Let's Encrypt or commercial CA)
- [ ] Set `APP_ENV=production`
- [ ] Configure AWS KMS for credential encryption
- [ ] Set up database backups with encryption
- [ ] Configure firewall rules (allow only necessary ports)
- [ ] Enable audit logging
- [ ] Set up monitoring and alerting
- [ ] Configure CORS for production domain only

### Environment Variables

Critical environment variables to set:

```bash
# Required
CREDENTIAL_MASTER_KEY=<base64-encoded-32-byte-key>
DB_PASSWORD=<strong-database-password>
REDIS_PASSWORD=<strong-redis-password>
KRATOS_WEBHOOK_SECRET=<random-secret-for-webhooks>

# Production recommended
CREDENTIAL_USE_KMS=true
CREDENTIAL_KMS_KEY_ID=<aws-kms-key-id>
DB_SSLMODE=require
SENTRY_ENABLED=true
SENTRY_DSN=<sentry-dsn>
TRACING_ENABLED=true
```

### Infrastructure Security

1. **Network Security**:
   - Use VPC with private subnets
   - NAT gateway for outbound traffic
   - Security groups with minimal permissions
   - No direct internet access to database

2. **Database Security**:
   - Enable encryption at rest
   - Force SSL/TLS connections
   - Regular automated backups
   - Point-in-time recovery enabled
   - Restricted access (whitelist IPs)

3. **Redis Security**:
   - Enable AUTH
   - Enable TLS
   - Bind to localhost or private network only
   - No public exposure

4. **Container Security** (if using Docker):
   - Use minimal base images (Alpine, Distroless)
   - Run as non-root user
   - Scan images for vulnerabilities
   - Keep images updated

## Incident Response

### Security Incident Procedure

1. **Detection**: Monitor logs for suspicious activity
2. **Containment**: Disable affected accounts/services
3. **Investigation**: Analyze logs and determine scope
4. **Eradication**: Remove threat and fix vulnerabilities
5. **Recovery**: Restore services and verify security
6. **Post-Incident**: Document and improve processes

### Emergency Contacts

Maintain a list of:
- Security team contacts
- Infrastructure/DevOps on-call
- Database administrator
- Legal/compliance team

### Data Breach Response

If a data breach occurs:

1. Immediately revoke compromised credentials
2. Notify affected users within 72 hours (GDPR requirement)
3. Document the incident
4. File reports as required by law
5. Implement additional security measures

## Compliance Considerations

### GDPR (General Data Protection Regulation)

- **Right to Access**: Users can request their data
- **Right to Erasure**: Implement data deletion
- **Data Portability**: Export user data in standard format
- **Data Protection**: Encryption and access controls
- **Breach Notification**: 72-hour notification requirement

### SOC 2 Type II

For SOC 2 compliance:

- Implement audit logging
- Access control reviews
- Vulnerability scanning
- Penetration testing
- Incident response procedures
- Data encryption
- Change management

## Security Testing

### Regular Security Activities

1. **Dependency Scanning**: Weekly automated scans
   ```bash
   govulncheck ./...
   npm audit
   ```

2. **Static Code Analysis**: On every commit
   ```bash
   gosec ./...
   ```

3. **Penetration Testing**: Annually (external firm)

4. **Security Code Review**: For all PRs touching:
   - Authentication/authorization code
   - Credential management
   - Encryption implementation
   - Input validation

## Contact and Reporting

### Security Vulnerability Reporting

If you discover a security vulnerability:

1. **DO NOT** open a public GitHub issue
2. Email: security@gorax.io (if available)
3. Use GitHub Security Advisories
4. Include:
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact
   - Suggested fix (if available)

### Bug Bounty Program

Details TBD

## Additional Resources

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [Ory Kratos Documentation](https://www.ory.sh/docs/kratos/)
- [AWS KMS Best Practices](https://docs.aws.amazon.com/kms/latest/developerguide/best-practices.html)
- [Go Security Policy](https://golang.org/security)

## Changelog

| Date | Version | Changes |
|------|---------|---------|
| 2025-12-20 | 1.0 | Initial security documentation |
