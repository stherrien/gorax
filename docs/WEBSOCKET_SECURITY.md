# WebSocket Origin Security

## Overview

The Gorax WebSocket implementation includes origin validation to prevent unauthorized connections from untrusted domains. This is a critical security measure to protect against Cross-Site WebSocket Hijacking (CSWSH) attacks.

## Configuration

WebSocket origin validation is configured via the `WEBSOCKET_ALLOWED_ORIGINS` environment variable.

### Environment Variable

```bash
# Single origin
WEBSOCKET_ALLOWED_ORIGINS=https://app.example.com

# Multiple origins (comma-separated)
WEBSOCKET_ALLOWED_ORIGINS=https://app.example.com,https://staging.example.com,http://localhost:3000

# Wildcard subdomain (matches any subdomain)
WEBSOCKET_ALLOWED_ORIGINS=https://*.example.com

# Wildcard port (matches any port on localhost)
WEBSOCKET_ALLOWED_ORIGINS=http://localhost:*

# Combination
WEBSOCKET_ALLOWED_ORIGINS=https://*.example.com,http://localhost:*
```

### Default Configuration

If `WEBSOCKET_ALLOWED_ORIGINS` is not set, the following defaults are used:

```
http://localhost:3000
http://localhost:5173
http://localhost:5174
```

These defaults are suitable for development but **MUST be changed for production**.

## Origin Matching Rules

### 1. Exact Match

The origin must match exactly, including protocol, host, and port.

```bash
# Configuration
WEBSOCKET_ALLOWED_ORIGINS=http://localhost:3000

# Allowed
ws://localhost:3000/api/v1/workflows/123/collaborate

# Denied
ws://localhost:5173/api/v1/workflows/123/collaborate  # Wrong port
wss://localhost:3000/api/v1/workflows/123/collaborate # Wrong protocol (wss vs ws)
```

### 2. Wildcard Subdomain

Use `*.domain.com` to match any subdomain (but not the base domain).

```bash
# Configuration
WEBSOCKET_ALLOWED_ORIGINS=https://*.example.com

# Allowed
wss://app.example.com/ws
wss://dev.app.example.com/ws
wss://staging.example.com/ws

# Denied
wss://example.com/ws              # Base domain not allowed
wss://app.different.com/ws        # Different domain
http://app.example.com/ws         # Wrong protocol
```

### 3. Wildcard Port

Use `host:*` to match any port on the specified host.

```bash
# Configuration
WEBSOCKET_ALLOWED_ORIGINS=http://localhost:*

# Allowed
ws://localhost:3000/ws
ws://localhost:5173/ws
ws://localhost:8080/ws

# Denied
ws://127.0.0.1:3000/ws            # Different host (localhost vs 127.0.0.1)
wss://localhost:3000/ws           # Wrong protocol
```

### 4. Protocol Matching

The protocol (http/https, ws/wss) must match exactly. WebSocket connections automatically use `ws://` for `http://` origins and `wss://` for `https://` origins.

```bash
# Configuration
WEBSOCKET_ALLOWED_ORIGINS=https://example.com

# Allowed
wss://example.com/ws              # HTTPS → WSS

# Denied
ws://example.com/ws               # HTTP → WS (protocol mismatch)
```

### 5. Case-Insensitive Domain

Domain names are matched case-insensitively.

```bash
# Configuration
WEBSOCKET_ALLOWED_ORIGINS=https://Example.COM

# Allowed
wss://example.com/ws
wss://EXAMPLE.COM/ws
wss://Example.com/ws
```

## Production Setup

### Recommended Production Configuration

```bash
# Use specific production domains
WEBSOCKET_ALLOWED_ORIGINS=https://app.example.com,https://app.example.co.uk

# Or use wildcard for subdomains
WEBSOCKET_ALLOWED_ORIGINS=https://*.example.com

# NEVER use in production:
# - http:// (unencrypted)
# - localhost
# - Wildcard ports (localhost:*)
# - Allow-all (empty config or *)
```

### Security Best Practices

1. **Always use HTTPS/WSS in production**
   ```bash
   # Good
   WEBSOCKET_ALLOWED_ORIGINS=https://app.example.com

   # Bad
   WEBSOCKET_ALLOWED_ORIGINS=http://app.example.com
   ```

2. **Be specific with origins**
   ```bash
   # Good - explicit domains
   WEBSOCKET_ALLOWED_ORIGINS=https://app.example.com,https://dashboard.example.com

   # Less secure - overly broad wildcard
   WEBSOCKET_ALLOWED_ORIGINS=https://*
   ```

3. **Don't include localhost in production**
   ```bash
   # Bad for production
   WEBSOCKET_ALLOWED_ORIGINS=https://app.example.com,http://localhost:3000
   ```

4. **Avoid wildcard ports in production**
   ```bash
   # Bad for production
   WEBSOCKET_ALLOWED_ORIGINS=https://app.example.com:*
   ```

## Development Setup

### Local Development

For local development, you can use wildcard ports for convenience:

```bash
# Allow any port on localhost
WEBSOCKET_ALLOWED_ORIGINS=http://localhost:*

# Or specify exact ports
WEBSOCKET_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:5173
```

### Development with Ngrok/Tunnel

If using ngrok or another tunneling service:

```bash
# Add your ngrok domain
WEBSOCKET_ALLOWED_ORIGINS=https://abc123.ngrok.io,http://localhost:3000
```

## Testing Origin Validation

### Using curl

```bash
# Test allowed origin
curl -i -N \
  -H "Connection: Upgrade" \
  -H "Upgrade: websocket" \
  -H "Sec-WebSocket-Version: 13" \
  -H "Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==" \
  -H "Origin: http://localhost:3000" \
  http://localhost:8080/api/v1/workflows/wf-123/collaborate

# Test denied origin
curl -i -N \
  -H "Connection: Upgrade" \
  -H "Upgrade: websocket" \
  -H "Sec-WebSocket-Version: 13" \
  -H "Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==" \
  -H "Origin: https://evil.com" \
  http://localhost:8080/api/v1/workflows/wf-123/collaborate
```

Expected responses:
- **Allowed origin**: `101 Switching Protocols` (after auth)
- **Denied origin**: `403 Forbidden` or connection refused

### Using JavaScript

```javascript
// Allowed origin (same domain)
const ws = new WebSocket('wss://app.example.com/api/v1/workflows/wf-123/collaborate');

ws.onopen = () => {
  console.log('Connected'); // Success
};

ws.onerror = (error) => {
  console.error('Connection failed:', error); // Should not happen for allowed origin
};
```

## Troubleshooting

### Connection Refused

**Symptom**: WebSocket connection fails with "403 Forbidden" or "Origin not allowed"

**Solution**: Check that your frontend origin is in `WEBSOCKET_ALLOWED_ORIGINS`:

```bash
# Check current origin in browser console
console.log(window.location.origin);

# Add to allowed origins
WEBSOCKET_ALLOWED_ORIGINS=https://your-actual-origin.com
```

### Localhost Connection Fails

**Symptom**: WebSocket connection fails when running locally

**Solution**: Ensure localhost is in allowed origins:

```bash
# Add localhost to allowed origins
WEBSOCKET_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:5173
```

### Wildcard Not Working

**Symptom**: Wildcard subdomain pattern doesn't match

**Solution**: Verify pattern syntax:

```bash
# Correct
WEBSOCKET_ALLOWED_ORIGINS=https://*.example.com

# Incorrect (missing protocol)
WEBSOCKET_ALLOWED_ORIGINS=*.example.com

# Incorrect (base domain won't match)
# *.example.com matches app.example.com but NOT example.com
```

### Multiple Origins Not Working

**Symptom**: Only first origin works

**Solution**: Ensure proper comma separation without spaces:

```bash
# Correct
WEBSOCKET_ALLOWED_ORIGINS=https://app.example.com,https://dashboard.example.com

# Incorrect (spaces will be trimmed, but be explicit)
WEBSOCKET_ALLOWED_ORIGINS=https://app.example.com, https://dashboard.example.com
```

## Security Considerations

### Why Origin Validation Matters

Without origin validation, an attacker could:

1. **Cross-Site WebSocket Hijacking**: Host a malicious website that connects to your WebSocket endpoints using the victim's credentials (cookies, tokens).
2. **Data Exfiltration**: Read sensitive real-time data through the WebSocket connection.
3. **Unauthorized Actions**: Send commands through the WebSocket on behalf of the victim.

### Defense in Depth

Origin validation is one layer of security. Also ensure:

1. **Authentication**: All WebSocket connections require valid authentication
2. **Authorization**: Users can only access resources they own
3. **HTTPS/WSS**: Use encrypted connections in production
4. **CSRF Tokens**: Include CSRF protection for WebSocket handshake
5. **Rate Limiting**: Prevent abuse of WebSocket connections

### Attack Scenarios

**Scenario 1: Missing Origin Validation**

```javascript
// Attacker's website (evil.com)
const ws = new WebSocket('wss://app.example.com/api/v1/workflows/wf-123/collaborate');
// Without origin validation, this would succeed and leak data
```

**Scenario 2: With Origin Validation**

```javascript
// Attacker's website (evil.com)
const ws = new WebSocket('wss://app.example.com/api/v1/workflows/wf-123/collaborate');
// Origin check fails → 403 Forbidden → Attack prevented
```

## Code Examples

### Backend Configuration (Go)

```go
// internal/config/websocket.go
config := config.WebSocketConfig{
    AllowedOrigins: []string{
        "https://app.example.com",
        "https://*.staging.example.com",
    },
}

upgrader := websocket.Upgrader{
    CheckOrigin: config.CheckOrigin(),
}
```

### Frontend Connection (TypeScript/React)

```typescript
// web/src/hooks/useCollaboration.ts
const connect = (workflowId: string) => {
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  const host = window.location.host;
  const url = `${protocol}//${host}/api/v1/workflows/${workflowId}/collaborate`;

  const ws = new WebSocket(url);

  ws.onopen = () => {
    console.log('Collaboration connected');
  };

  ws.onerror = (error) => {
    console.error('Collaboration connection failed:', error);
  };

  return ws;
};
```

## Migration Guide

### Upgrading from Unrestricted Origins

If you're upgrading from a version without origin validation:

1. **Audit your frontend origins**:
   ```bash
   # List all domains where your frontend is deployed
   - Production: https://app.example.com
   - Staging: https://staging.example.com
   - Development: http://localhost:3000
   ```

2. **Set environment variable**:
   ```bash
   # Production
   WEBSOCKET_ALLOWED_ORIGINS=https://app.example.com

   # Staging
   WEBSOCKET_ALLOWED_ORIGINS=https://staging.example.com

   # Development
   WEBSOCKET_ALLOWED_ORIGINS=http://localhost:*
   ```

3. **Deploy and test**:
   - Deploy backend with new configuration
   - Test WebSocket connections from all frontend deployments
   - Monitor logs for "origin not allowed" errors

4. **Rollback plan**:
   - If issues occur, temporarily use wildcard:
     ```bash
     WEBSOCKET_ALLOWED_ORIGINS=https://*.example.com,http://localhost:*
     ```
   - Investigate and tighten restrictions once stable

## References

- [OWASP WebSocket Security](https://owasp.org/www-community/vulnerabilities/WebSocket_security)
- [RFC 6455: The WebSocket Protocol](https://tools.ietf.org/html/rfc6455)
- [gorilla/websocket CheckOrigin](https://pkg.go.dev/github.com/gorilla/websocket#Upgrader)
