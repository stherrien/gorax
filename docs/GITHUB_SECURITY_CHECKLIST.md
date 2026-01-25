# GitHub Security Checklist

Comprehensive security checklist for GitHub repository configuration, focusing on protecting the gorax repository and ensuring compliance with security best practices.

## Table of Contents

1. [Access Control](#access-control)
2. [Authentication & Authorization](#authentication--authorization)
3. [Branch Protection](#branch-protection)
4. [Secrets Management](#secrets-management)
5. [Code Scanning & Security](#code-scanning--security)
6. [Dependency Management](#dependency-management)
7. [Audit & Compliance](#audit--compliance)
8. [Incident Response](#incident-response)
9. [Regular Security Reviews](#regular-security-reviews)

---

## Access Control

### Repository Access

- [ ] **Minimum necessary access**: Users have least privilege access
- [ ] **Role-based access control**: Appropriate roles assigned (Read, Triage, Write, Maintain, Admin)
- [ ] **No unnecessary admins**: Limit admin access to 2-3 people
- [ ] **Team-based access**: Use GitHub teams instead of individual collaborators where possible
- [ ] **Regular access review**: Review collaborators quarterly

### Organization Settings (if applicable)

- [ ] **2FA required**: All organization members must enable 2FA
- [ ] **SAML SSO enabled**: Use Single Sign-On for enterprise accounts
- [ ] **IP allowlist**: Configure IP restrictions if applicable
- [ ] **OAuth app restrictions**: Limit which OAuth apps can access repositories

### Verification Commands

```bash
# List all collaborators
gh api repos/{owner}/{repo}/collaborators | jq -r '.[] | "\(.login): \(.permissions)"'

# Check if 2FA is enforced (organization level)
gh api orgs/{org} | jq '.two_factor_requirement_enabled'

# List organization members without 2FA (if admin)
gh api orgs/{org}/members?filter=2fa_disabled
```

---

## Authentication & Authorization

### Personal Access Tokens

- [ ] **Use fine-grained tokens**: Prefer fine-grained PATs over classic
- [ ] **Token expiration**: Set expiration dates (max 90 days)
- [ ] **Minimum scopes**: Tokens have only necessary permissions
- [ ] **Token rotation**: Regular rotation schedule established
- [ ] **Token audit**: Review active tokens quarterly

### Deploy Keys & SSH Keys

- [ ] **Read-only deploy keys**: Use read-only where possible
- [ ] **Key rotation**: Rotate SSH keys annually
- [ ] **Ed25519 keys**: Use Ed25519 instead of RSA for new keys
- [ ] **Key monitoring**: Track deploy key usage

### Verification Commands

```bash
# List deploy keys
gh api repos/{owner}/{repo}/keys

# List SSH keys for user
gh api user/keys

# Check token permissions (for current token)
gh api user | jq '.permissions'
```

---

## Branch Protection

### Main Branch

- [ ] **Pull request required**: Cannot push directly to main
- [ ] **Review required**: At least 1 approval required
- [ ] **Dismiss stale reviews**: Enabled
- [ ] **Code owner review**: Required (if CODEOWNERS exists)
- [ ] **Status checks required**: All CI/CD checks must pass
- [ ] **Up-to-date branch required**: Branch must be current with base
- [ ] **Signed commits**: Required (recommended)
- [ ] **Linear history**: Required (prevents merge commits)
- [ ] **Force push disabled**: Cannot rewrite history
- [ ] **Deletion disabled**: Cannot delete branch
- [ ] **Admin enforcement**: Rules apply to administrators
- [ ] **Restrict pushes**: Only specific users/teams can push

### Dev Branch

- [ ] **Pull request required**: Cannot push directly to dev
- [ ] **Review required**: At least 1 approval required
- [ ] **Status checks required**: Critical checks must pass
- [ ] **Force push disabled**: Cannot rewrite history
- [ ] **Deletion disabled**: Cannot delete branch

### Release/Hotfix Branches (if applicable)

- [ ] **Protection configured**: Appropriate rules for release branches
- [ ] **Review requirements**: Higher approval count for releases

### Verification Commands

```bash
# Check main branch protection
gh api repos/{owner}/{repo}/branches/main/protection | jq '.'

# Verify required status checks
gh api repos/{owner}/{repo}/branches/main/protection | jq '.required_status_checks.contexts'

# Check if force push is disabled
gh api repos/{owner}/{repo}/branches/main/protection | jq '.allow_force_pushes.enabled'
# Should return: false

# Check if signed commits required
gh api repos/{owner}/{repo}/branches/main/protection | jq '.required_signatures.enabled'
```

---

## Secrets Management

### Secret Configuration

- [ ] **No secrets in code**: No hardcoded secrets in repository
- [ ] **GitHub Secrets used**: All secrets in GitHub Secrets, not code
- [ ] **Environment-specific secrets**: Production secrets isolated
- [ ] **Secret rotation schedule**: Regular rotation policy established
- [ ] **Minimal secret exposure**: Secrets limited to necessary workflows
- [ ] **No logging secrets**: Workflows don't log secret values

### Secret Security

- [ ] **Strong secrets**: All secrets meet complexity requirements
  - Database passwords: ≥32 characters
  - API tokens: ≥48 characters
  - Encryption keys: ≥32 bytes (base64 encoded)
- [ ] **Unique secrets**: Different secrets for staging/production
- [ ] **Secure generation**: Cryptographically secure random generation
- [ ] **Backup storage**: Secrets backed up in secure vault (1Password, Vault, etc.)
- [ ] **Access logging**: Track who views/modifies secrets

### Required Secrets Audit

Repository Secrets:
- [ ] `CODECOV_TOKEN` (optional)

Staging Environment:
- [ ] `STAGING_URL`
- [ ] `STAGING_DB_PASSWORD`
- [ ] `STAGING_REDIS_PASSWORD`
- [ ] `STAGING_HEALTH_CHECK_TOKEN`

Production Environment:
- [ ] `PRODUCTION_URL`
- [ ] `PRODUCTION_DB_PASSWORD`
- [ ] `PRODUCTION_REDIS_PASSWORD`
- [ ] `HEALTH_CHECK_TOKEN`
- [ ] `CREDENTIAL_MASTER_KEY`

Infrastructure (choose one):
- [ ] Kubernetes: `KUBECONFIG`, `K8S_TOKEN`, etc.
- [ ] AWS ECS: `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, etc.
- [ ] SSH: `SSH_PRIVATE_KEY`, `SSH_HOST`, etc.

### Verification Commands

```bash
# List repository secrets (names only)
gh secret list

# List environment secrets
gh api repos/{owner}/{repo}/environments/production/secrets | jq -r '.secrets[].name'

# Check for secrets in code (run from repo root)
git grep -i "password\|secret\|token\|key" | grep -v ".md:" | grep -v "template"

# Scan for leaked secrets
gitleaks detect --source . --verbose
```

---

## Code Scanning & Security

### Automated Security Scanning

- [ ] **CodeQL enabled**: CodeQL analysis running on PRs
- [ ] **Secret scanning enabled**: GitHub secret scanning active
- [ ] **Push protection enabled**: Prevents committing secrets
- [ ] **Dependency scanning**: Dependabot alerts enabled
- [ ] **Container scanning**: Docker images scanned for vulnerabilities
- [ ] **gosec enabled**: Go security scanning in CI
- [ ] **npm audit enabled**: JavaScript dependency audit in CI

### Security Policies

- [ ] **SECURITY.md exists**: Vulnerability reporting process documented
- [ ] **Security contact**: Security team contact information available
- [ ] **Response SLA**: Vulnerability response time documented
- [ ] **Disclosure policy**: Responsible disclosure policy defined

### Code Review Security

- [ ] **Security-focused reviews**: Reviewers trained on security
- [ ] **Security checklist**: Reviewers use security checklist for PRs
- [ ] **High-risk change review**: Security team reviews critical changes
- [ ] **External dependency review**: New dependencies reviewed for security

### Verification Commands

```bash
# Check if secret scanning is enabled
gh api repos/{owner}/{repo} | jq '.security_and_analysis.secret_scanning.status'
# Should return: "enabled"

# Check if push protection is enabled
gh api repos/{owner}/{repo} | jq '.security_and_analysis.secret_scanning_push_protection.status'
# Should return: "enabled"

# Check Dependabot status
gh api repos/{owner}/{repo}/automated-security-fixes

# List CodeQL alerts
gh api repos/{owner}/{repo}/code-scanning/alerts

# List Dependabot alerts
gh api repos/{owner}/{repo}/dependabot/alerts
```

---

## Dependency Management

### Dependency Security

- [ ] **Dependabot enabled**: Automated dependency updates configured
- [ ] **Auto-merge safe updates**: Automatic merging of patch updates
- [ ] **Review major updates**: Manual review for major version updates
- [ ] **License compliance**: Dependencies have acceptable licenses
- [ ] **No known vulnerabilities**: All critical/high vulnerabilities patched

### Dependabot Configuration

- [ ] **Schedule configured**: `dependabot.yml` configured
- [ ] **Version updates enabled**: Automated version updates
- [ ] **Security updates enabled**: Automated security fixes
- [ ] **Auto-merge rules**: Safe updates auto-merged
- [ ] **Update limits**: Reasonable pull request limits set

### Package Registry Security

- [ ] **Package verification**: Verify package integrity
- [ ] **Checksums verified**: Dependency checksums checked
- [ ] **Signed packages**: Prefer signed/verified packages
- [ ] **Private registry**: Use private registry for internal packages

### Verification Commands

```bash
# Check if Dependabot is configured
cat .github/dependabot.yml

# List Dependabot alerts
gh api repos/{owner}/{repo}/dependabot/alerts | jq -r '.[] | "\(.security_advisory.severity): \(.security_advisory.summary)"'

# Check for outdated dependencies (Go)
go list -u -m all

# Check for outdated dependencies (npm)
cd web && npm outdated

# Audit dependencies (Go)
go list -json -m all | nancy sleuth

# Audit dependencies (npm)
cd web && npm audit
```

---

## Audit & Compliance

### Audit Logging

- [ ] **Audit log enabled**: Organization audit log enabled (if applicable)
- [ ] **Log retention**: Audit logs retained for required period
- [ ] **Log monitoring**: Automated monitoring of suspicious activity
- [ ] **Deployment logging**: All deployments logged with approver
- [ ] **Secret access logging**: Track secret modifications

### Compliance Requirements

- [ ] **Change management**: All production changes require PR + approval
- [ ] **Deployment approval**: Production requires designated approver
- [ ] **Rollback capability**: Documented rollback procedure
- [ ] **Backup policy**: Regular backups of critical data
- [ ] **Documentation**: Runbooks and procedures documented

### Access Audits

- [ ] **Quarterly access review**: Review all repository access
- [ ] **Remove inactive users**: Remove users who haven't contributed in 6 months
- [ ] **Review admin access**: Verify admin access is still required
- [ ] **Audit service accounts**: Review deploy keys, tokens, service accounts

### Verification Commands

```bash
# List recent repository events
gh api repos/{owner}/{repo}/events | jq -r '.[] | "\(.type): \(.actor.login)"'

# View audit log (organization)
gh api orgs/{org}/audit-log | jq -r '.[] | "\(.created_at): \(.action) by \(.actor)"'

# List collaborators with last activity
gh api repos/{owner}/{repo}/collaborators | jq -r '.[] | "\(.login): \(.permissions)"'

# List webhooks (could be security risk)
gh api repos/{owner}/{repo}/hooks | jq -r '.[] | "\(.id): \(.config.url)"'
```

---

## Incident Response

### Incident Preparation

- [ ] **Incident response plan**: Documented IR procedure
- [ ] **On-call rotation**: Defined on-call schedule
- [ ] **Communication channels**: Incident channels configured (Slack, PagerDuty)
- [ ] **Escalation path**: Clear escalation procedures
- [ ] **Contact list**: Updated emergency contacts

### Secret Compromise Response

- [ ] **Rotation procedure**: Quick secret rotation procedure documented
- [ ] **Impact assessment**: Process to assess compromise impact
- [ ] **Notification procedure**: Know who to notify for compromised secrets
- [ ] **Forensics capability**: Ability to investigate compromise

### Security Incident Types

**Compromised Credentials:**
1. Immediately rotate affected secrets
2. Review access logs for unauthorized access
3. Assess what data was accessed
4. Notify affected parties if needed
5. Document incident and lessons learned

**Malicious Code Injection:**
1. Immediately revert malicious commits
2. Review all recent commits by compromised account
3. Rotate all credentials that account had access to
4. Review branch protection (how did it get merged?)
5. Notify security team and stakeholders

**Unauthorized Access:**
1. Immediately revoke compromised access
2. Review audit logs for activity
3. Change all secrets accessible by that account
4. Assess data exfiltration risk
5. Document and improve access controls

### Verification & Testing

```bash
# Test secret rotation (staging)
# 1. Generate new secret
NEW_SECRET=$(openssl rand -base64 32)

# 2. Update GitHub secret
gh secret set STAGING_DB_PASSWORD --env staging --body "$NEW_SECRET"

# 3. Trigger deployment to use new secret
gh workflow run deploy-staging.yml

# 4. Verify application works with new secret
curl -f https://staging.gorax.dev/health/db
```

---

## Regular Security Reviews

### Weekly Tasks

- [ ] Review Dependabot alerts and create fixes
- [ ] Review new pull requests for security issues
- [ ] Monitor CodeQL findings
- [ ] Check for any secret scanning alerts

### Monthly Tasks

- [ ] Rotate staging environment secrets
- [ ] Review access logs for anomalies
- [ ] Update security documentation if needed
- [ ] Test incident response procedures (tabletop exercise)
- [ ] Review and triage security alerts

### Quarterly Tasks

- [ ] Full access audit (collaborators, teams, tokens)
- [ ] Review branch protection rules
- [ ] Audit webhooks and integrations
- [ ] Update and test incident response plan
- [ ] Review and update security policies
- [ ] Security training for team members
- [ ] Penetration testing (if applicable)

### Annual Tasks

- [ ] Rotate production environment secrets
- [ ] Rotate SSH keys and deploy keys
- [ ] Full security audit by external team
- [ ] Update security certifications
- [ ] Review compliance requirements
- [ ] Disaster recovery drill

### Review Checklist Template

```markdown
## Security Review - [Date]

### Access Control
- [ ] Reviewed all collaborators
- [ ] Removed inactive users
- [ ] Verified 2FA for all users
- [ ] Audited admin access

### Secrets
- [ ] All secrets up to date
- [ ] Rotation schedule followed
- [ ] No secrets in code
- [ ] Backup verified

### Branch Protection
- [ ] Main branch protected
- [ ] Dev branch protected
- [ ] All checks configured
- [ ] No unauthorized changes

### Security Scanning
- [ ] CodeQL running
- [ ] Secret scanning active
- [ ] Dependabot enabled
- [ ] No high-severity alerts

### Incident Response
- [ ] IR plan up to date
- [ ] Contact list current
- [ ] Runbooks reviewed
- [ ] Team trained

### Notes
[Any findings or actions taken]

### Action Items
- [ ] [Action item 1]
- [ ] [Action item 2]

**Reviewed by:** [Name]
**Next review:** [Date]
```

---

## Automated Security Checks

### GitHub Actions Workflow

Create `.github/workflows/security-audit.yml`:

```yaml
name: Security Audit

on:
  schedule:
    - cron: '0 0 * * 1'  # Weekly on Monday
  workflow_dispatch:

jobs:
  security-audit:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Run validation script
        run: ./scripts/validate-github-config.sh --json > audit-report.json

      - name: Check for secrets in code
        run: |
          git grep -i "password\|secret\|token\|key" | grep -v ".md:" | grep -v "template" && exit 1 || exit 0

      - name: Dependency audit (Go)
        run: |
          go list -json -m all | nancy sleuth

      - name: Dependency audit (npm)
        working-directory: ./web
        run: npm audit --production

      - name: Upload audit report
        uses: actions/upload-artifact@v4
        with:
          name: security-audit-report
          path: audit-report.json
          retention-days: 90

      - name: Notify on failure
        if: failure()
        run: |
          echo "Security audit failed! Review audit report."
          # Add Slack/email notification here
```

---

## Security Tools & Resources

### Recommended Tools

**Secret Scanning:**
- [gitleaks](https://github.com/zricethezav/gitleaks) - Scan for secrets in git history
- [truffleHog](https://github.com/trufflesecurity/trufflehog) - Find secrets in code

**Dependency Scanning:**
- [nancy](https://github.com/sonatype-nexus-community/nancy) - Go dependency scanner
- [npm audit](https://docs.npmjs.com/cli/v8/commands/npm-audit) - npm vulnerability scanner
- [Snyk](https://snyk.io) - Comprehensive dependency scanner

**Code Scanning:**
- [gosec](https://github.com/securego/gosec) - Go security scanner
- [ESLint Security Plugin](https://github.com/nodesecurity/eslint-plugin-security) - JavaScript security linter
- [CodeQL](https://codeql.github.com) - Semantic code analysis

**Container Scanning:**
- [Trivy](https://github.com/aquasecurity/trivy) - Container vulnerability scanner
- [Grype](https://github.com/anchore/grype) - Vulnerability scanner

### Security Resources

- [GitHub Security Best Practices](https://docs.github.com/en/code-security)
- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [CIS Benchmarks](https://www.cisecurity.org/cis-benchmarks/)
- [NIST Cybersecurity Framework](https://www.nist.gov/cyberframework)

---

## Quick Commands Reference

```bash
# Full security validation
./scripts/validate-github-config.sh --verbose

# Scan for secrets in code
gitleaks detect --source . --verbose

# Check for vulnerabilities
go list -json -m all | nancy sleuth
cd web && npm audit

# List all repository secrets
gh secret list

# Check branch protection
gh api repos/{owner}/{repo}/branches/main/protection

# Review recent commits
git log --all --pretty=format:'%h %an %ae %s' --since='1 week ago'

# Check for unsigned commits
git log --show-signature | grep -B 2 "No signature"

# Audit collaborator access
gh api repos/{owner}/{repo}/collaborators

# Generate security report
./scripts/validate-github-config.sh --json > security-report.json
```

---

## Compliance Frameworks

### SOC 2

- [ ] Access control policies documented
- [ ] Change management process followed
- [ ] Audit logging enabled and retained
- [ ] Incident response plan documented
- [ ] Regular security reviews conducted

### GDPR

- [ ] Data processing documented
- [ ] Access controls in place
- [ ] Audit trails maintained
- [ ] Breach notification procedure ready
- [ ] Data retention policies defined

### PCI DSS (if handling payment data)

- [ ] Strong access control measures
- [ ] Network security maintained
- [ ] Regular security monitoring
- [ ] Information security policies documented
- [ ] Regular vulnerability assessments

---

## Appendix

### Contact Information

**Security Team:**
- Email: security@gorax.dev
- Slack: #gorax-security
- PagerDuty: @gorax-security-oncall

**Incident Response:**
- On-call: @gorax-oncall
- Escalation: [Manager name]
- 24/7 Hotline: [Phone number]

### Related Documentation

- [GitHub Configuration Guide](./GITHUB_CONFIGURATION_GUIDE.md)
- [Security Documentation](./SECURITY.md)
- [Security Audit Report](./SECURITY_AUDIT_REPORT.md)
- [Deployment Guide](./DEPLOYMENT.md)
- [Incident Response Plan](./INCIDENT_RESPONSE.md)

### Changelog

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0.0 | 2026-01-02 | @stherrien | Initial security checklist |

---

**Document Version:** 1.0.0
**Last Updated:** 2026-01-02
**Next Review:** 2026-04-02
**Owner:** Security Team (@gorax-team/security)
