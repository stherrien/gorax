# GitHub Repository Setup - Quick Start

Quick reference guide for setting up GitHub repository configuration for the gorax project.

## TL;DR - Fastest Setup

```bash
# 1. Install prerequisites
brew install gh jq terraform

# 2. Authenticate
gh auth login

# 3. Run automated setup
./scripts/github-setup.sh --interactive

# 4. Verify configuration
./scripts/validate-github-config.sh

# Done! ✓
```

---

## What Gets Configured

### Environments
- **staging** - Auto-deploy from `dev` branch
- **production** - Requires approval, deploys from `main` branch

### Branch Protection
- **main** - Strict protection (1 approval, all checks, signed commits)
- **dev** - Standard protection (1 approval, critical checks)

### Secrets
- Repository secrets (CODECOV_TOKEN)
- Environment secrets (URLs, passwords, tokens)
- Infrastructure secrets (Kubernetes/AWS/SSH)

### Repository Settings
- Merge strategies (squash enabled, rebase disabled)
- Auto-delete branches
- Security features (Dependabot, secret scanning)

---

## Three Ways to Setup

### Option 1: Automated Script (Recommended)

**Pros:** Fast, interactive prompts, idempotent
**Best for:** Quick setup, testing

```bash
# Dry run first
./scripts/github-setup.sh --dry-run

# Interactive setup (prompts for values)
./scripts/github-setup.sh --interactive

# Non-interactive with environment variables
PRODUCTION_REVIEWER_ID=12345678 \
STAGING_DB_PASSWORD="$(openssl rand -base64 32)" \
./scripts/github-setup.sh
```

### Option 2: Terraform (Production)

**Pros:** Infrastructure as code, version controlled, repeatable
**Best for:** Production deployments, team environments

```bash
cd terraform/github

# Configure
cp terraform.tfvars.example terraform.tfvars
vim terraform.tfvars  # Add your values

# Apply
terraform init
terraform plan
terraform apply
```

### Option 3: Manual (Understanding)

**Pros:** Full control, understand each step
**Best for:** Learning, troubleshooting

Follow the detailed guide:
- [GitHub Configuration Guide](./GITHUB_CONFIGURATION_GUIDE.md)

---

## Prerequisites

### Required Tools

```bash
# GitHub CLI
brew install gh
gh --version  # >= 2.0.0

# Terraform (if using IaC)
brew install terraform
terraform --version  # >= 1.0.0

# jq (JSON processor)
brew install jq
jq --version  # >= 1.6
```

### Required Access

- Repository **Admin** access
- GitHub account with **2FA enabled**
- Personal Access Token with `repo` and `admin:org` scopes

### Authentication

```bash
# Authenticate with GitHub CLI
gh auth login

# Verify
gh auth status

# Set default repository
gh repo set-default stherrien/gorax
```

---

## Step-by-Step Setup

### 1. Environments Setup

```bash
# Create staging environment
gh api repos/stherrien/gorax/environments/staging -X PUT \
  -f 'deployment_branch_policy[protected_branches]=false' \
  -f 'deployment_branch_policy[custom_branch_policies]=true'

# Allow dev branch deployments
gh api repos/stherrien/gorax/environments/staging/deployment-branch-policies -X POST \
  -f 'name=dev' -f 'type=branch'

# Create production environment with approval
USER_ID=$(gh api user | jq .id)
gh api repos/stherrien/gorax/environments/production -X PUT \
  -f "reviewers[0][type]=User" \
  -f "reviewers[0][id]=$USER_ID" \
  -f 'wait_timer=5'

# Allow main branch deployments
gh api repos/stherrien/gorax/environments/production/deployment-branch-policies -X POST \
  -f 'name=main' -f 'type=branch'
```

### 2. Secrets Setup

```bash
# Repository secrets
gh secret set CODECOV_TOKEN --body "your-token"

# Staging secrets
gh secret set STAGING_URL --env staging --body "https://staging.gorax.dev"
gh secret set STAGING_DB_PASSWORD --env staging --body "$(openssl rand -base64 32)"

# Production secrets
gh secret set PRODUCTION_URL --env production --body "https://gorax.dev"
gh secret set PRODUCTION_DB_PASSWORD --env production --body "$(openssl rand -base64 48)"
gh secret set HEALTH_CHECK_TOKEN --env production --body "Bearer $(openssl rand -hex 48)"
```

### 3. Branch Protection Setup

```bash
# Main branch protection
gh api repos/stherrien/gorax/branches/main/protection -X PUT \
  --input .github/branch-protection/main.json

# Dev branch protection
gh api repos/stherrien/gorax/branches/dev/protection -X PUT \
  --input .github/branch-protection/dev.json
```

### 4. Verification

```bash
# Run validation script
./scripts/validate-github-config.sh --verbose

# Check specific components
./scripts/validate-github-config.sh --environments
./scripts/validate-github-config.sh --secrets
./scripts/validate-github-config.sh --branch-protection
```

---

## Common Tasks

### Add a Secret

```bash
# Repository secret
gh secret set SECRET_NAME --body "secret-value"

# Environment secret
gh secret set SECRET_NAME --env production --body "secret-value"

# From file
gh secret set SSH_PRIVATE_KEY < ~/.ssh/id_ed25519

# Interactive (won't echo)
gh secret set API_TOKEN
# Paste value and press Ctrl+D
```

### Add Production Reviewer

```bash
# Get user ID
USER_ID=$(gh api users/{username} | jq .id)

# Add as reviewer
gh api repos/stherrien/gorax/environments/production -X PUT \
  -f "reviewers[0][type]=User" \
  -f "reviewers[0][id]=$USER_ID"
```

### Update Branch Protection

```bash
# View current protection
gh api repos/stherrien/gorax/branches/main/protection

# Update specific setting (example: required approvals)
gh api repos/stherrien/gorax/branches/main/protection -X PUT \
  --input - <<'EOF'
{
  "required_pull_request_reviews": {
    "required_approving_review_count": 2,
    "dismiss_stale_reviews": true
  },
  "required_status_checks": null,
  "enforce_admins": true
}
EOF
```

---

## Troubleshooting

### Issue: "Resource not accessible by integration"

**Solution:** Check workflow permissions:
```bash
# Settings → Actions → General → Workflow permissions
# Select: "Read and write permissions"
```

### Issue: Environment not found

**Solution:** Verify environment exists:
```bash
gh api repos/stherrien/gorax/environments | jq -r '.environments[].name'
```

### Issue: Secret not available in workflow

**Solutions:**
1. Check secret name is exact (case-sensitive)
2. For environment secrets, ensure job specifies environment:
   ```yaml
   jobs:
     deploy:
       environment: production  # Required!
   ```
3. Verify secret is set:
   ```bash
   gh secret list
   gh api repos/stherrien/gorax/environments/production/secrets
   ```

### Issue: Cannot merge PR

**Possible causes:**
1. Branch not up to date → Update branch
2. Missing required status check → Wait for checks
3. No approval → Request review
4. Conversations not resolved → Resolve comments

---

## Security Best Practices

### DO

- ✅ Use environment-specific secrets
- ✅ Rotate secrets regularly (90 days)
- ✅ Use strong, random passwords (≥32 chars)
- ✅ Enable 2FA for all users
- ✅ Review access quarterly
- ✅ Enable all security features
- ✅ Monitor Dependabot alerts

### DON'T

- ❌ Commit secrets to code
- ❌ Share secrets between environments
- ❌ Use weak passwords
- ❌ Disable branch protection
- ❌ Skip PR reviews
- ❌ Ignore security alerts
- ❌ Give unnecessary admin access

---

## Verification Checklist

After setup, verify everything works:

- [ ] Staging environment exists and accessible
- [ ] Production environment exists with reviewers
- [ ] All required secrets are set
- [ ] Main branch is protected (cannot push directly)
- [ ] Dev branch is protected
- [ ] Status checks are configured
- [ ] Dependabot is enabled
- [ ] Secret scanning is enabled
- [ ] Can create PR and run checks
- [ ] Can deploy to staging (auto)
- [ ] Can deploy to production (with approval)

**Run automated verification:**
```bash
./scripts/validate-github-config.sh
```

---

## Next Steps

1. **Test the setup:**
   ```bash
   # Create test PR
   git checkout -b test/verify-setup
   echo "test" >> README.md
   git commit -am "test: verify GitHub configuration"
   git push origin test/verify-setup
   gh pr create --title "Test: Verify Setup" --body "Testing GitHub configuration"

   # Watch checks run
   gh pr checks --watch

   # Clean up
   gh pr close --delete-branch
   ```

2. **Configure team access:**
   - Add collaborators
   - Create teams (if organization)
   - Set up CODEOWNERS file

3. **Set up monitoring:**
   - Configure Slack notifications
   - Set up Sentry for error tracking
   - Enable production monitoring

4. **Document for team:**
   - Share configuration guide with team
   - Train team on security practices
   - Establish incident response procedures

---

## Quick Reference

### File Locations

```
gorax/
├── .github/
│   ├── environments/
│   │   ├── staging.json          # Staging config template
│   │   └── production.json       # Production config template
│   └── secrets.template.yml      # All secrets documented
├── docs/
│   ├── GITHUB_CONFIGURATION_GUIDE.md    # Comprehensive guide
│   ├── GITHUB_SECURITY_CHECKLIST.md    # Security checklist
│   └── GITHUB_SETUP_QUICK_START.md     # This file
├── scripts/
│   ├── github-setup.sh           # Automated setup
│   └── validate-github-config.sh # Validation script
└── terraform/github/             # Terraform IaC
    ├── main.tf
    ├── environments.tf
    ├── branch-protection.tf
    ├── secrets.tf
    ├── variables.tf
    └── README.md
```

### Important Commands

```bash
# Environments
gh api repos/{owner}/{repo}/environments

# Secrets
gh secret list
gh secret set SECRET_NAME
gh secret delete SECRET_NAME

# Branch Protection
gh api repos/{owner}/{repo}/branches/{branch}/protection

# Validation
./scripts/validate-github-config.sh

# Setup
./scripts/github-setup.sh --interactive
```

### Useful Links

- [GitHub Configuration Guide](./GITHUB_CONFIGURATION_GUIDE.md) - Detailed setup
- [Security Checklist](./GITHUB_SECURITY_CHECKLIST.md) - Security best practices
- [Terraform README](../terraform/github/README.md) - IaC documentation
- [Secrets Template](../.github/secrets.template.yml) - Required secrets
- [GitHub Docs](https://docs.github.com) - Official documentation

---

## Support

**Questions?**
- Check [Troubleshooting](#troubleshooting) section above
- Review [GitHub Configuration Guide](./GITHUB_CONFIGURATION_GUIDE.md)
- Ask in Slack: `#gorax-devops`
- Open an issue with `security` label

**Emergency?**
- PagerDuty: @gorax-devops-oncall
- Slack: `#gorax-incidents`
- Email: devops@gorax.dev

---

**Last Updated:** 2026-01-02
**Version:** 1.0.0
**Maintained By:** DevOps Team
