# SSRF Protection Implementation

## Overview

Server-Side Request Forgery (SSRF) protection has been implemented for Gorax HTTP workflow actions to prevent attackers from exploiting workflow configurations to access internal resources or metadata services.

## What is SSRF?

SSRF is a vulnerability where an attacker can force a server to make requests to unintended locations, typically internal services or cloud metadata endpoints. In the context of workflow automation, this could allow malicious workflows to:

- Access internal APIs and services (192.168.x.x, 10.x.x.x)
- Read cloud provider metadata (AWS: 169.254.169.254, GCP, Azure)
- Access localhost services (127.0.0.1, ::1)
- Scan internal networks
- Bypass firewalls and access controls

## Implementation Details

### Architecture

The SSRF protection is implemented in three layers:

1. **URL Validator** (`internal/security/url_validator.go`)
   - Validates URL schemes (only http/https allowed)
   - Resolves DNS to check final IP addresses
   - Checks IPs against blocklist/allowlist

2. **HTTP Action Integration** (`internal/executor/actions/http.go`)
   - Validates URLs before making requests
   - Applies validation after interpolation
   - Returns clear error messages

3. **Configuration** (`internal/config/config.go`)
   - Environment-based configuration
   - Tenant-specific allow/block lists (optional)
   - Enabled by default for security

### Blocked Ranges

By default, the following IP ranges are blocked:

**IPv4:**
- `127.0.0.0/8` - Loopback addresses (localhost)
- `10.0.0.0/8` - Private network (RFC 1918)
- `172.16.0.0/12` - Private network (RFC 1918)
- `192.168.0.0/16` - Private network (RFC 1918)
- `169.254.0.0/16` - Link-local (AWS/GCP metadata service)

**IPv6:**
- `::1/128` - Loopback
- `fc00::/7` - Unique local addresses (private)
- `fe80::/10` - Link-local addresses

### Blocked Schemes

Only `http://` and `https://` schemes are allowed. The following are blocked:
- `file://` - Local file access
- `ftp://` - FTP protocol
- `gopher://` - Gopher protocol
- `data://` - Data URLs
- `javascript://` - JavaScript execution
- Any other non-HTTP(S) scheme

## Configuration

### Environment Variables

```bash
# Enable/disable SSRF protection (default: true)
SSRF_PROTECTION_ENABLED=true

# Allow specific internal networks (comma-separated CIDR ranges)
# Example: Allow internal API gateway
SSRF_ALLOWED_NETWORKS=192.168.1.0/24,10.0.0.0/8

# Block additional networks beyond defaults (comma-separated CIDR ranges)
# Example: Block specific public range
SSRF_BLOCKED_NETWORKS=203.0.113.0/24
```

### Production Recommendations

1. **Keep SSRF protection enabled** - Set `SSRF_PROTECTION_ENABLED=true`
2. **Minimize allowed networks** - Only allowlist specific internal services if absolutely necessary
3. **Use DNS-based access control** - Configure internal services with DNS names that resolve to public IPs
4. **Log blocked attempts** - Monitor SSRF protection logs for potential attacks
5. **Regular audits** - Review workflow configurations for suspicious URL patterns

### Development Environment

For local development with test servers:
```bash
# Option 1: Disable SSRF protection (not recommended)
SSRF_PROTECTION_ENABLED=false

# Option 2: Allowlist localhost (better for testing specific scenarios)
SSRF_ALLOWED_NETWORKS=127.0.0.0/8
```

## Usage Examples

### Allowed Request

```json
{
  "type": "action:http",
  "config": {
    "method": "POST",
    "url": "https://api.example.com/webhook",
    "body": {"data": "{{trigger.payload}}"}
  }
}
```

This will succeed because `api.example.com` resolves to a public IP address.

### Blocked Request (Loopback)

```json
{
  "type": "action:http",
  "config": {
    "method": "GET",
    "url": "http://127.0.0.1:8080/admin"
  }
}
```

**Error:** `SSRF protection: blocked IP address: 127.0.0.1`

### Blocked Request (AWS Metadata)

```json
{
  "type": "action:http",
  "config": {
    "method": "GET",
    "url": "http://169.254.169.254/latest/meta-data/"
  }
}
```

**Error:** `SSRF protection: blocked IP address: 169.254.169.254`

### Blocked Request (Private Network)

```json
{
  "type": "action:http",
  "config": {
    "method": "GET",
    "url": "http://192.168.1.100/api"
  }
}
```

**Error:** `SSRF protection: blocked IP address: 192.168.1.100`

### Allowed with Allowlist

If you need to call internal services, configure an allowlist:

```bash
SSRF_ALLOWED_NETWORKS=192.168.1.0/24
```

Then this request will succeed:
```json
{
  "type": "action:http",
  "config": {
    "method": "GET",
    "url": "http://192.168.1.100/api"
  }
}
```

## Security Considerations

### DNS Rebinding

The current implementation resolves DNS at validation time. However, DNS rebinding attacks could theoretically bypass this by:
1. Returning a public IP during validation
2. Returning a private IP during the actual request

**Mitigation:**
- Use short DNS TTLs and re-validate periodically (future enhancement)
- Implement connection-time IP validation (future enhancement)
- Monitor and alert on suspicious DNS patterns

### URL Encoding

The validator checks URLs after interpolation but before encoding. Attackers might try:
- URL encoding: `http://127.0.0.1` → `http://127.0.0.1` (blocked correctly)
- Decimal notation: `http://2130706433/` → Blocked correctly
- Hex notation: `http://0x7f.0x0.0x0.0x1/` → Blocked correctly

The Go `net` package handles these variations correctly.

### Redirects

The HTTP client follows redirects by default. An attacker could:
1. Host a public URL that redirects to a private IP
2. Bypass initial validation

**Mitigation:**
- Implement redirect validation (future enhancement)
- Limit redirect count (currently limited to 10)
- Validate each redirect destination

### Time-of-Check to Time-of-Use (TOCTOU)

There's a small window between DNS resolution and the actual HTTP request where DNS could change.

**Mitigation:**
- Accept this small risk for performance
- Monitor for suspicious patterns
- Implement connection-time validation if needed (future enhancement)

## Testing

### Unit Tests

The implementation includes comprehensive unit tests:

**URL Validator Tests** (`internal/security/url_validator_test.go`):
- Valid URLs
- Blocked schemes (file://, ftp://, etc.)
- Loopback addresses (127.x.x.x, ::1)
- Private IPs (10.x.x.x, 172.16.x.x, 192.168.x.x)
- Link-local (169.254.x.x - AWS/GCP metadata)
- IPv6 private ranges
- DNS resolution
- Edge cases (URL encoding, case sensitivity)
- Configuration options

**HTTP Action Integration Tests** (`internal/executor/actions/http_test.go`):
- SSRF blocking in HTTP actions
- Integration with URL interpolation
- Error message validation

### Running Tests

```bash
# Run all SSRF-related tests
go test ./internal/security/... ./internal/executor/actions/... -v

# Run with race detector
go test ./internal/security/... ./internal/executor/actions/... -race

# Run specific SSRF tests
go test ./internal/executor/actions/... -v -run TestHTTPAction_SSRFProtection
```

### Manual Testing

Test in a staging environment:

```bash
# 1. Verify loopback is blocked
curl -X POST http://localhost:8080/api/workflows/execute \
  -H "Content-Type: application/json" \
  -d '{
    "workflow_id": "test",
    "steps": [{
      "type": "action:http",
      "config": {"method": "GET", "url": "http://127.0.0.1/admin"}
    }]
  }'
# Expected: Error with "SSRF protection: blocked IP address"

# 2. Verify AWS metadata is blocked
curl -X POST http://localhost:8080/api/workflows/execute \
  -H "Content-Type: application/json" \
  -d '{
    "workflow_id": "test",
    "steps": [{
      "type": "action:http",
      "config": {"method": "GET", "url": "http://169.254.169.254/latest/meta-data/"}
    }]
  }'
# Expected: Error with "SSRF protection: blocked IP address"

# 3. Verify public URLs work
curl -X POST http://localhost:8080/api/workflows/execute \
  -H "Content-Type: application/json" \
  -d '{
    "workflow_id": "test",
    "steps": [{
      "type": "action:http",
      "config": {"method": "GET", "url": "https://httpbin.org/get"}
    }]
  }'
# Expected: Success
```

## Monitoring

### Logging

SSRF protection violations are logged with context:

```go
// Use ValidateURLWithLogging for audit trails
validator.ValidateURLWithLogging(url, func(msg string, fields map[string]interface{}) {
    log.WithFields(fields).Warn(msg)
})
```

Example log entry:
```json
{
  "level": "warn",
  "msg": "SSRF protection blocked URL",
  "url": "http://169.254.169.254/latest/meta-data/",
  "error": "blocked IP address: 169.254.169.254",
  "tenant_id": "tenant-123",
  "workflow_id": "wf-456",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### Metrics

Consider adding Prometheus metrics:

```go
ssrf_blocked_total{reason="loopback|private|metadata|scheme"}
ssrf_allowed_total
ssrf_validation_duration_seconds
```

## Future Enhancements

### Short-term
1. Add redirect URL validation
2. Implement connection-time IP validation
3. Add telemetry for blocked attempts
4. Create admin dashboard for SSRF events

### Long-term
1. Machine learning-based anomaly detection
2. Tenant-specific security policies
3. Integration with threat intelligence feeds
4. Webhook URL verification before workflow execution

## References

- [OWASP SSRF Prevention Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Server_Side_Request_Forgery_Prevention_Cheat_Sheet.html)
- [PortSwigger SSRF](https://portswigger.net/web-security/ssrf)
- [AWS SSRF Protection](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/instancedata-data-retrieval.html)
- [RFC 1918 - Address Allocation for Private Internets](https://tools.ietf.org/html/rfc1918)
- [RFC 3986 - Uniform Resource Identifier (URI): Generic Syntax](https://tools.ietf.org/html/rfc3986)

## Support

For questions or issues related to SSRF protection:
1. Check the logs for detailed error messages
2. Review this documentation
3. Test with `SSRF_PROTECTION_ENABLED=false` to isolate issues
4. File a GitHub issue with reproduction steps
