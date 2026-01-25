# Okta SAML 2.0 Setup Guide

## Overview

This guide walks through configuring Okta as a SAML 2.0 identity provider for Gorax.

## Prerequisites

- Okta admin account
- Gorax tenant admin access
- Domain ownership verification

## Step 1: Create SAML Application in Okta

### 1.1 Navigate to Applications

1. Log in to Okta Admin Console
2. Go to **Applications** → **Applications**
3. Click **Create App Integration**

### 1.2 Configure Application

1. Select **SAML 2.0**
2. Click **Next**

### 1.3 General Settings

- **App name**: `Gorax`
- **App logo**: (optional)
- Click **Next**

### 1.4 SAML Settings

#### Single Sign-On URL (ACS URL)
```
https://app.gorax.io/api/v1/sso/acs
```

- ✅ Check "Use this for Recipient URL and Destination URL"

#### Audience URI (SP Entity ID)
```
https://app.gorax.io
```

#### Default RelayState
```
(leave blank or use custom state)
```

#### Name ID Format
```
EmailAddress
```

#### Application Username
```
Email
```

#### Update Application Username On
```
Create and update
```

### 1.5 Attribute Statements

Add these attribute mappings:

| Name | Name Format | Value |
|------|-------------|-------|
| `email` | Unspecified | `user.email` |
| `firstName` | Unspecified | `user.firstName` |
| `lastName` | Unspecified | `user.lastName` |

### 1.6 Group Attribute Statements (Optional)

For group-based access control:

| Name | Name Format | Filter | Value |
|------|-------------|--------|-------|
| `groups` | Unspecified | Matches regex | `.*` |

### 1.7 Finish Setup

1. Click **Next**
2. Select "I'm an Okta customer adding an internal app"
3. Click **Finish**

## Step 2: Get IdP Metadata

### 2.1 Download Metadata

1. In the application page, go to **Sign On** tab
2. Under **SAML 2.0**, find **Metadata URL**
3. Right-click **Identity Provider metadata** and copy link
4. The URL will look like:
   ```
   https://company.okta.com/app/exk.../sso/saml/metadata
   ```

### 2.2 Alternative: Copy Metadata XML

If you prefer to upload XML directly:

1. Click **View SAML setup instructions**
2. Copy the entire **IDP metadata** XML

## Step 3: Assign Users

### 3.1 Assign People

1. Go to **Assignments** tab
2. Click **Assign** → **Assign to People**
3. Select users who should have access
4. Click **Assign** and **Done**

### 3.2 Assign Groups (Optional)

1. Click **Assign** → **Assign to Groups**
2. Select groups
3. Click **Assign** and **Done**

## Step 4: Configure Gorax

### 4.1 Create SSO Provider in Gorax

Using the API:

```bash
curl -X POST https://app.gorax.io/api/v1/sso/providers \
  -H "Authorization: Bearer $YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Okta SAML",
    "provider_type": "saml",
    "enabled": true,
    "enforce_sso": false,
    "domains": ["yourcompany.com"],
    "config": {
      "entity_id": "https://app.gorax.io",
      "acs_url": "https://app.gorax.io/api/v1/sso/acs",
      "idp_metadata_url": "https://company.okta.com/app/exk.../sso/saml/metadata",
      "attribute_mapping": {
        "email": "email",
        "first_name": "firstName",
        "last_name": "lastName",
        "groups": "groups"
      },
      "sign_authn_requests": false
    }
  }'
```

### 4.2 Or Using Admin UI

1. Log in to Gorax as tenant admin
2. Navigate to **Settings** → **SSO**
3. Click **Add SSO Provider**
4. Select **SAML 2.0**
5. Fill in the form:
   - **Name**: `Okta SAML`
   - **IdP Metadata URL**: Paste the metadata URL from Okta
   - **Domains**: `yourcompany.com`
   - **Attribute Mapping**: Use default or customize
6. Click **Create**

## Step 5: Test SSO

### 5.1 Test from Okta

1. In Okta, go to your Gorax application
2. Click **View App** button
3. Should redirect to Gorax and log you in

### 5.2 Test from Gorax

1. Go to Gorax login page
2. Enter email: `user@yourcompany.com`
3. Click **Continue with SSO**
4. Should redirect to Okta
5. Authenticate in Okta
6. Should redirect back to Gorax and be logged in

## Step 6: Advanced Configuration

### Enable Request Signing (Optional)

If Okta requires signed requests:

#### 6.1 Generate Certificate

```bash
# Generate private key
openssl genrsa -out saml.key 2048

# Generate certificate
openssl req -new -x509 -key saml.key -out saml.crt -days 3650
```

#### 6.2 Upload to Okta

1. In Okta application, go to **Sign On** tab
2. Click **Edit** under **SAML 2.0**
3. Upload the `saml.crt` certificate

#### 6.3 Update Gorax Config

```json
{
  "config": {
    "sign_authn_requests": true,
    "certificate": "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----",
    "private_key": "-----BEGIN RSA PRIVATE KEY-----\n...\n-----END RSA PRIVATE KEY-----"
  }
}
```

### Enable SSO Enforcement

Force all users in domain to use SSO:

```bash
curl -X PUT https://app.gorax.io/api/v1/sso/providers/{provider-id} \
  -H "Authorization: Bearer $YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "enforce_sso": true
  }'
```

## Troubleshooting

### Issue: "Invalid signature"

**Cause**: Certificate mismatch

**Solution**:
1. Download fresh IdP metadata
2. Update provider configuration
3. Test again

### Issue: "User not found"

**Cause**: User not assigned to application

**Solution**:
1. In Okta, assign user to Gorax application
2. Try logging in again

### Issue: "Attribute not found"

**Cause**: Attribute mapping mismatch

**Solution**:
1. Check attribute names in Okta application
2. Update attribute mapping in Gorax
3. Test again

### Issue: "RelayState invalid"

**Cause**: State management issue

**Solution**:
1. Clear browser cookies
2. Try IdP-initiated flow (from Okta)
3. Check session storage configuration

## Security Best Practices

1. **MFA**: Enable MFA in Okta for all users
2. **Session Duration**: Set appropriate session timeout in Okta
3. **Certificate Rotation**: Rotate SAML certificates annually
4. **Audit Logs**: Regularly review SSO login events
5. **Group Mapping**: Use groups for role-based access control

## Support

- Okta Documentation: https://help.okta.com
- Gorax Support: support@gorax.io
