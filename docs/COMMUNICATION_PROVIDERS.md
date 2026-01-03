# Communication Providers Guide

This guide covers email and SMS communication providers for Gorax workflow automation.

## Table of Contents

- [Overview](#overview)
- [Email Providers](#email-providers)
  - [SendGrid](#sendgrid)
  - [Mailgun](#mailgun)
  - [AWS SES](#aws-ses)
  - [SMTP](#smtp)
- [SMS Providers](#sms-providers)
  - [Twilio](#twilio)
  - [AWS SNS](#aws-sns)
  - [MessageBird](#messagebird)
- [Workflow Actions](#workflow-actions)
- [Configuration](#configuration)
- [Credentials](#credentials)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)

## Overview

Gorax supports multiple email and SMS providers for sending communications from workflows. All providers support:

- Individual message sending
- Bulk message sending
- Error handling and retry logic
- Message tracking and logging
- Secure credential management

### Supported Providers

**Email Providers:**
- SendGrid (transactional email)
- Mailgun (email service)
- AWS SES (Amazon Simple Email Service)
- SMTP (generic SMTP servers)

**SMS Providers:**
- Twilio (SMS and voice)
- AWS SNS (Amazon Simple Notification Service)
- MessageBird (international SMS)

## Email Providers

### SendGrid

**Features:**
- High deliverability rates
- Advanced analytics
- Template support
- Marketing campaigns

**Setup:**

1. Sign up at [SendGrid](https://sendgrid.com/)
2. Create an API key with "Mail Send" permissions
3. Verify sender identity (email or domain)

**Credential Configuration:**

```json
{
  "name": "SendGrid Production",
  "type": "email_sendgrid",
  "value": {
    "api_key": "SG.your-api-key-here"
  }
}
```

**Workflow Action Example:**

```json
{
  "type": "send_email",
  "config": {
    "provider": "sendgrid",
    "credential_id": "your-credential-id",
    "from": "noreply@yourdomain.com",
    "to": ["user@example.com"],
    "subject": "Welcome to Our Service",
    "body": "Thank you for signing up!",
    "body_html": "<h1>Welcome!</h1><p>Thank you for signing up!</p>"
  }
}
```

**Rate Limits:**
- Free: 100 emails/day
- Essential: 40,000 emails/month
- Pro: 100,000+ emails/month

**Cost:**
- Free tier available
- Paid plans from $19.95/month

### Mailgun

**Features:**
- Email validation
- Detailed analytics
- Webhook support
- European data residency options

**Setup:**

1. Sign up at [Mailgun](https://www.mailgun.com/)
2. Add and verify your domain
3. Create an API key

**Credential Configuration:**

```json
{
  "name": "Mailgun Production",
  "type": "email_mailgun",
  "value": {
    "domain": "mg.yourdomain.com",
    "api_key": "your-mailgun-api-key"
  }
}
```

**Workflow Action Example:**

```json
{
  "type": "send_email",
  "config": {
    "provider": "mailgun",
    "credential_id": "your-credential-id",
    "from": "support@yourdomain.com",
    "to": ["customer@example.com"],
    "cc": ["team@yourdomain.com"],
    "subject": "Order Confirmation",
    "body_html": "<p>Your order has been confirmed.</p>",
    "attachments": [
      {
        "filename": "receipt.pdf",
        "content": "base64-encoded-content",
        "content_type": "application/pdf"
      }
    ]
  }
}
```

**Rate Limits:**
- Free: 5,000 emails/month (3 months)
- Flex: Pay as you go
- Foundation: 50,000 emails/month

**Cost:**
- Free tier: 5,000 emails for 3 months
- Paid: $35/month and up

### AWS SES

**Features:**
- Cost-effective at scale
- High throughput
- Integration with AWS ecosystem
- Dedicated IP addresses available

**Setup:**

1. Enable AWS SES in your AWS account
2. Verify email addresses or domains
3. Request production access (initially in sandbox mode)
4. Configure AWS credentials

**Credential Configuration:**

```json
{
  "name": "AWS SES Production",
  "type": "email_aws_ses",
  "value": {
    "region": "us-east-1",
    "access_key_id": "AKIAIOSFODNN7EXAMPLE",
    "secret_access_key": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
  }
}
```

**Workflow Action Example:**

```json
{
  "type": "send_email",
  "config": {
    "provider": "aws_ses",
    "credential_id": "your-credential-id",
    "from": "notifications@yourdomain.com",
    "to": ["user@example.com"],
    "subject": "Password Reset Request",
    "body": "Click the link to reset your password: {{reset_link}}",
    "reply_to": "support@yourdomain.com"
  }
}
```

**Rate Limits:**
- Sandbox: 200 emails/day, 1 email/second
- Production: 50,000 emails/day initially (can be increased)

**Cost:**
- $0.10 per 1,000 emails
- First 62,000 emails/month free (with EC2)
- Very cost-effective at scale

### SMTP

**Features:**
- Works with any SMTP server
- No vendor lock-in
- Self-hosted options

**Supported Services:**
- Gmail
- Outlook/Office 365
- Custom mail servers
- Any RFC 5321 compliant SMTP server

**Setup:**

1. Enable SMTP access in your email provider
2. Generate app-specific password (for Gmail)
3. Note server hostname and port

**Credential Configuration:**

```json
{
  "name": "Gmail SMTP",
  "type": "email_smtp",
  "value": {
    "username": "your-email@gmail.com",
    "password": "your-app-specific-password"
  }
}
```

**Workflow Action Example:**

```json
{
  "type": "send_email",
  "config": {
    "provider": "smtp",
    "credential_id": "your-credential-id",
    "smtp_config": {
      "host": "smtp.gmail.com",
      "port": 587,
      "use_tls": true
    },
    "from": "your-email@gmail.com",
    "to": ["recipient@example.com"],
    "subject": "Test Email",
    "body": "This is a test email sent via SMTP."
  }
}
```

**Common SMTP Settings:**

| Provider | Host | Port | TLS |
|----------|------|------|-----|
| Gmail | smtp.gmail.com | 587 | Yes |
| Outlook | smtp-mail.outlook.com | 587 | Yes |
| Office 365 | smtp.office365.com | 587 | Yes |
| Yahoo | smtp.mail.yahoo.com | 587 | Yes |

**Rate Limits:**
- Gmail: 500 emails/day (free), 2,000/day (Google Workspace)
- Varies by provider

## SMS Providers

### Twilio

**Features:**
- Global SMS coverage
- Programmable voice
- WhatsApp Business API
- Detailed analytics

**Setup:**

1. Sign up at [Twilio](https://www.twilio.com/)
2. Get your Account SID and Auth Token
3. Purchase a phone number

**Credential Configuration:**

```json
{
  "name": "Twilio Production",
  "type": "sms_twilio",
  "value": {
    "account_sid": "ACxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
    "auth_token": "your-auth-token"
  }
}
```

**Workflow Action Example:**

```json
{
  "type": "send_sms",
  "config": {
    "provider": "twilio",
    "credential_id": "your-credential-id",
    "from": "+1234567890",
    "to": "+19876543210",
    "message": "Your verification code is: 123456"
  }
}
```

**Rate Limits:**
- 10 segments per second per phone number
- Higher limits available on request

**Cost:**
- US: $0.0079/SMS
- International: $0.01-$0.50/SMS (varies by country)
- Phone numbers: $1-$2/month

### AWS SNS

**Features:**
- Pay-per-use pricing
- Global coverage
- No phone number required
- Integration with AWS ecosystem

**Setup:**

1. Enable AWS SNS in your AWS account
2. Configure AWS credentials with SNS permissions
3. Ensure sender ID is configured for your region

**Credential Configuration:**

```json
{
  "name": "AWS SNS Production",
  "type": "sms_aws_sns",
  "value": {
    "region": "us-east-1",
    "access_key_id": "AKIAIOSFODNN7EXAMPLE",
    "secret_access_key": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
  }
}
```

**Workflow Action Example:**

```json
{
  "type": "send_sms",
  "config": {
    "provider": "aws_sns",
    "credential_id": "your-credential-id",
    "from": "YourApp",
    "to": "+19876543210",
    "message": "Your order has been shipped. Track it here: {{tracking_url}}"
  }
}
```

**Rate Limits:**
- Default: 20 messages/second
- Can be increased via AWS Support

**Cost:**
- US: $0.00645/SMS
- International: $0.00645-$0.70/SMS (varies)
- No monthly fees

### MessageBird

**Features:**
- 217 countries coverage
- Omnichannel messaging
- Number verification
- Two-factor authentication

**Setup:**

1. Sign up at [MessageBird](https://www.messagebird.com/)
2. Get your API key from the dashboard
3. Purchase a virtual number (optional)

**Credential Configuration:**

```json
{
  "name": "MessageBird Production",
  "type": "sms_messagebird",
  "value": {
    "api_key": "your-messagebird-api-key"
  }
}
```

**Workflow Action Example:**

```json
{
  "type": "send_sms",
  "config": {
    "provider": "messagebird",
    "credential_id": "your-credential-id",
    "from": "YourBrand",
    "to": "+31612345678",
    "message": "Hello from MessageBird! Visit: {{website_url}}"
  }
}
```

**Rate Limits:**
- 50 messages/second default
- Higher limits available

**Cost:**
- Varies by destination
- Europe: €0.06-€0.10/SMS
- Competitive international rates

## Workflow Actions

### Send Email Action

**Action Type:** `send_email`

**Configuration Fields:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| provider | string | Yes | Email provider (sendgrid, mailgun, aws_ses, smtp) |
| credential_id | string | Yes | ID of the credential to use |
| from | string | Yes | Sender email address |
| to | array | Yes | Recipient email addresses |
| cc | array | No | CC recipients |
| bcc | array | No | BCC recipients |
| subject | string | Yes | Email subject |
| body | string | No* | Plain text body |
| body_html | string | No* | HTML body |
| reply_to | string | No | Reply-to address |
| headers | object | No | Custom email headers |
| attachments | array | No | Email attachments |
| smtp_config | object | No** | SMTP configuration |

\* Either `body` or `body_html` is required
\** Required for SMTP provider

**Output:**

```json
{
  "success": true,
  "message_id": "abc123xyz",
  "status": "sent",
  "sent_at": "2024-01-15T10:30:00Z"
}
```

### Send SMS Action

**Action Type:** `send_sms`

**Configuration Fields:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| provider | string | Yes | SMS provider (twilio, aws_sns, messagebird) |
| credential_id | string | Yes | ID of the credential to use |
| from | string | Yes | Sender phone number or ID |
| to | string | Yes | Recipient phone number (E.164 format) |
| message | string | Yes | SMS message content (max 1600 chars) |

**Output:**

```json
{
  "success": true,
  "message_id": "SM123abc",
  "status": "sent",
  "cost": 0.0079,
  "sent_at": "2024-01-15T10:30:00Z"
}
```

## Configuration

### Environment Variables

Add these to your `.env` file:

```bash
# SendGrid
SENDGRID_API_KEY=your-api-key

# Mailgun
MAILGUN_DOMAIN=mg.yourdomain.com
MAILGUN_API_KEY=your-api-key

# AWS SES
AWS_SES_REGION=us-east-1
AWS_SES_ACCESS_KEY_ID=your-access-key
AWS_SES_SECRET_ACCESS_KEY=your-secret-key

# SMTP
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-password
SMTP_USE_TLS=true

# Twilio
TWILIO_ACCOUNT_SID=your-account-sid
TWILIO_AUTH_TOKEN=your-auth-token
TWILIO_PHONE_NUMBER=+1234567890

# AWS SNS
AWS_SNS_REGION=us-east-1
AWS_SNS_ACCESS_KEY_ID=your-access-key
AWS_SNS_SECRET_ACCESS_KEY=your-secret-key

# MessageBird
MESSAGEBIRD_API_KEY=your-api-key
```

## Credentials

### Creating Credentials

Use the Credentials API to create communication provider credentials:

**SendGrid Example:**

```bash
curl -X POST http://localhost:8080/api/credentials \
  -H "Content-Type: application/json" \
  -d '{
    "name": "SendGrid Production",
    "type": "email_sendgrid",
    "description": "Production email sending via SendGrid",
    "value": {
      "api_key": "SG.your-api-key"
    }
  }'
```

**Twilio Example:**

```bash
curl -X POST http://localhost:8080/api/credentials \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Twilio Production",
    "type": "sms_twilio",
    "description": "Production SMS via Twilio",
    "value": {
      "account_sid": "ACxxxxx",
      "auth_token": "your-token"
    }
  }'
```

### Credential Types

| Type | Provider | Required Fields |
|------|----------|----------------|
| email_sendgrid | SendGrid | api_key |
| email_mailgun | Mailgun | domain, api_key |
| email_aws_ses | AWS SES | region, access_key_id, secret_access_key |
| email_smtp | SMTP | username, password |
| sms_twilio | Twilio | account_sid, auth_token |
| sms_aws_sns | AWS SNS | region, access_key_id, secret_access_key |
| sms_messagebird | MessageBird | api_key |

## Best Practices

### Email

1. **Verify Domains:** Always verify sender domains to improve deliverability
2. **Authentication:** Use SPF, DKIM, and DMARC records
3. **Warm Up IPs:** Gradually increase sending volume with new IPs
4. **List Management:** Maintain clean recipient lists, remove bounces
5. **Unsubscribe Links:** Include unsubscribe links in marketing emails
6. **Attachments:** Keep total attachment size under 10MB
7. **Templates:** Use templates for consistent branding
8. **Testing:** Test emails across different clients

### SMS

1. **Number Format:** Use E.164 format (+1234567890)
2. **Message Length:** Keep under 160 characters for single segment
3. **Opt-In:** Ensure recipients opted in to receive messages
4. **Opt-Out:** Honor opt-out requests immediately
5. **Timing:** Respect time zones and quiet hours
6. **Clear Sender:** Use recognizable sender IDs
7. **Shortlinks:** Use URL shorteners for links
8. **Cost Monitoring:** Monitor SMS costs, especially for international

### Security

1. **Credential Rotation:** Rotate API keys regularly
2. **Least Privilege:** Use minimum required permissions
3. **Audit Logs:** Enable audit logging for compliance
4. **Rate Limiting:** Implement rate limiting to prevent abuse
5. **Input Validation:** Validate all recipient data
6. **PII Protection:** Handle personal information securely
7. **Encryption:** Ensure TLS/SSL for all communications

### Error Handling

1. **Retry Logic:** Implement exponential backoff for retries
2. **Dead Letter Queue:** Handle failed messages appropriately
3. **Monitoring:** Set up alerts for high failure rates
4. **Fallback Providers:** Configure backup providers
5. **Status Tracking:** Track message delivery status
6. **Error Logging:** Log errors with context for debugging

## Troubleshooting

### Common Email Issues

**Problem:** Emails going to spam

**Solutions:**
- Verify sender domain with SPF/DKIM/DMARC
- Warm up new IP addresses gradually
- Avoid spam trigger words in subject/body
- Maintain low bounce/complaint rates
- Use authenticated sender addresses

**Problem:** High bounce rate

**Solutions:**
- Validate email addresses before sending
- Remove hard bounces from lists
- Check for typos in email addresses
- Verify recipient domains exist

**Problem:** API authentication errors

**Solutions:**
- Verify API key is correct and active
- Check credential hasn't expired
- Ensure proper permissions for API key
- Verify account is in good standing

### Common SMS Issues

**Problem:** Messages not delivering

**Solutions:**
- Verify phone number format (E.164)
- Check recipient is in supported country
- Ensure sufficient account balance
- Verify sender ID is approved for region
- Check for carrier blocking

**Problem:** High costs

**Solutions:**
- Monitor international vs domestic sends
- Use local numbers for better rates
- Consider bulk pricing plans
- Optimize message length
- Review and adjust sending patterns

**Problem:** Rate limit errors

**Solutions:**
- Implement request queuing
- Use bulk send APIs where available
- Contact provider to increase limits
- Distribute load across multiple numbers

### Debugging

**Enable Debug Logging:**

```bash
LOG_LEVEL=debug
```

**Check Communication Events:**

```sql
SELECT *
FROM communication_events
WHERE tenant_id = 'your-tenant-id'
ORDER BY sent_at DESC
LIMIT 100;
```

**Monitor Failed Messages:**

```sql
SELECT
  provider,
  status,
  COUNT(*) as count,
  MAX(error_message) as last_error
FROM communication_events
WHERE status = 'failed'
  AND sent_at > NOW() - INTERVAL '24 hours'
GROUP BY provider, status;
```

## Support

For provider-specific issues:

- **SendGrid:** https://support.sendgrid.com/
- **Mailgun:** https://help.mailgun.com/
- **AWS SES:** https://docs.aws.amazon.com/ses/
- **Twilio:** https://support.twilio.com/
- **AWS SNS:** https://docs.aws.amazon.com/sns/
- **MessageBird:** https://support.messagebird.com/

For Gorax issues:
- GitHub Issues: https://github.com/stherrien/gorax/issues
- Documentation: https://github.com/stherrien/gorax/docs
