# GitHub Repository Configuration Guide

Complete guide for configuring GitHub repository settings including environments, secrets, branch protection, and CI/CD integration for the gorax project.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Quick Start](#quick-start)
3. [Manual Configuration](#manual-configuration)
4. [Automated Setup](#automated-setup)
5. [Environment Configuration](#environment-configuration)
6. [Secrets Management](#secrets-management)
7. [Branch Protection Rules](#branch-protection-rules)
8. [Repository Settings](#repository-settings)
9. [Verification](#verification)
10. [Troubleshooting](#troubleshooting)
11. [Maintenance](#maintenance)

---

## Prerequisites

### Required Access

- **Repository Admin** access to configure settings
- **Organization Owner** access (if configuring organization-level settings)
- GitHub account with 2FA enabled (required for sensitive operations)

### Required Tools

Install these tools before proceeding:

```bash
# GitHub CLI
brew install gh

# Terraform (for automated setup)
brew install terraform

# jq (for JSON parsing)
brew install jq

# Verify installations
gh --version    # Should be >= 2.0.0
terraform --version  # Should be >= 1.0.0
jq --version    # Should be >= 1.6
```

### Initial Authentication

```bash
# Authenticate with GitHub CLI
gh auth login

# Verify authentication
gh auth status

# Set repository context
gh repo set-default stherrien/gorax
```

---

## Quick Start

For a fully automated setup, use the provided script:

```bash
# Review what will be configured
./scripts/github-setup.sh --dry-run

# Run the automated setup
./scripts/github-setup.sh

# Or use interactive mode
./scripts/github-setup.sh --interactive

# Verify configuration
./scripts/validate-github-config.sh
```

For Terraform-based setup:

```bash
cd terraform/github
terraform init
terraform plan
terraform apply
```

---

## Manual Configuration

### Step-by-Step Web Interface Configuration

Follow these steps if you prefer manual setup through GitHub's web interface.

#### 1. Access Repository Settings

1. Navigate to your repository: `https://github.com/stherrien/gorax`
2. Click **Settings** tab
3. Ensure you have admin access (you'll see all configuration options)

#### 2. Configure General Settings

Navigate to: **Settings → General**

**Repository Details:**
- Description: `Enterprise workflow automation platform with AI-powered workflow generation`
- Website: `https://gorax.dev`
- Topics: `workflow-automation`, `golang`, `react`, `typescript`, `low-code`

**Features:**
- ✅ Issues
- ✅ Projects
- ❌ Wiki (use docs/ instead)
- ✅ Discussions
- ✅ Sponsorships

**Pull Requests:**
- ✅ Allow squash merging
- ✅ Default to pull request title and commit details
- ✅ Allow merge commits
- ❌ Allow rebase merging (to maintain linear history)
- ✅ Always suggest updating pull request branches
- ✅ Automatically delete head branches

**Archives:**
- ❌ Include Git LFS objects in archives

---

## Environment Configuration

### Overview

Environments provide deployment protection rules and environment-specific secrets.

### Staging Environment

Navigate to: **Settings → Environments → New environment**

**Configuration:**

| Setting | Value |
|---------|-------|
| Name | `staging` |
| Environment URL | `https://staging.gorax.dev` |
| Wait timer | None |
| Required reviewers | None (auto-deploy) |
| Deployment branches | `dev` branch only |

**Protection Rules:**

1. Click **Configure environment**
2. **Deployment branches:**
   - Select "Selected branches"
   - Add pattern: `dev`
3. **Environment secrets:**
   - Add secrets specific to staging (see [Secrets Management](#secrets-management))

**Environment Variables (if needed):**
```
ENVIRONMENT=staging
LOG_LEVEL=debug
```

### Production Environment

Navigate to: **Settings → Environments → New environment**

**Configuration:**

| Setting | Value |
|---------|-------|
| Name | `production` |
| Environment URL | `https://gorax.dev` |
| Wait timer | 5 minutes (optional) |
| Required reviewers | At least 1 |
| Deployment branches | `main` branch only |

**Protection Rules:**

1. Click **Configure environment**
2. **Required reviewers:**
   - ✅ Enable required reviewers
   - Add at least 1 team member (e.g., @stherrien)
   - Consider adding a team (e.g., @gorax-team/platform-engineers)
3. **Wait timer:**
   - ✅ Enable wait timer
   - Set to **5 minutes** (allows time to cancel if needed)
4. **Deployment branches:**
   - Select "Selected branches"
   - Add pattern: `main`
5. **Environment secrets:**
   - Add secrets specific to production

**Environment Variables:**
```
ENVIRONMENT=production
LOG_LEVEL=info
```

### Development Environment (Optional)

For testing deployment workflows locally:

| Setting | Value |
|---------|-------|
| Name | `development` |
| Environment URL | `http://localhost:8080` |
| Protection rules | None |

---

## Secrets Management

### Required Secrets

Navigate to: **Settings → Secrets and variables → Actions**

#### Repository Secrets

These secrets are available to all workflows:

| Secret Name | Description | How to Obtain | Example |
|------------|-------------|---------------|---------|
| `CODECOV_TOKEN` | Codecov upload token (optional) | 1. Go to [codecov.io](https://codecov.io)<br>2. Add repository<br>3. Copy token | `abc123def456...` |
| `GITHUB_TOKEN` | Automatically provided | Built-in, no setup needed | N/A |

#### Staging Environment Secrets

Navigate to: **Settings → Environments → staging → Add secret**

| Secret Name | Description | Example |
|------------|-------------|---------|
| `STAGING_URL` | Staging environment URL | `https://staging.gorax.dev` |
| `STAGING_DB_PASSWORD` | Staging database password | Strong password |
| `STAGING_REDIS_PASSWORD` | Staging Redis password | Strong password |
| `STAGING_HEALTH_CHECK_TOKEN` | Token for health checks | `Bearer sk_staging_...` |

#### Production Environment Secrets

Navigate to: **Settings → Environments → production → Add secret**

| Secret Name | Description | Example |
|------------|-------------|---------|
| `PRODUCTION_URL` | Production environment URL | `https://gorax.dev` |
| `PRODUCTION_DB_PASSWORD` | Production database password | Strong password |
| `PRODUCTION_REDIS_PASSWORD` | Production Redis password | Strong password |
| `HEALTH_CHECK_TOKEN` | Token for health checks | `Bearer sk_prod_...` |
| `CREDENTIAL_MASTER_KEY` | Encryption master key | 32-byte base64 string |

### Infrastructure-Specific Secrets

Choose the relevant set based on your deployment infrastructure:

#### Option A: Kubernetes Deployment

| Secret Name | Description | How to Obtain |
|------------|-------------|---------------|
| `KUBECONFIG` | Base64-encoded kubeconfig | `cat ~/.kube/config \| base64` |
| `K8S_CLUSTER_URL` | Kubernetes API URL | From your K8s provider |
| `K8S_TOKEN` | Service account token | `kubectl get secret ...` |
| `K8S_NAMESPACE` | Deployment namespace | `gorax-production` |

#### Option B: AWS ECS Deployment

| Secret Name | Description | How to Obtain |
|------------|-------------|---------------|
| `AWS_ACCESS_KEY_ID` | AWS access key | AWS IAM Console |
| `AWS_SECRET_ACCESS_KEY` | AWS secret key | AWS IAM Console |
| `AWS_REGION` | AWS region | `us-east-1` |
| `ECS_CLUSTER` | ECS cluster name | `gorax-production` |
| `ECS_SERVICE` | ECS service name | `gorax-api` |
| `ECS_TASK_DEFINITION` | Task definition family | `gorax-api` |

#### Option C: SSH Deployment

| Secret Name | Description | How to Obtain |
|------------|-------------|---------------|
| `SSH_PRIVATE_KEY` | Private SSH key | `cat ~/.ssh/id_ed25519` |
| `SSH_HOST` | Server hostname/IP | `prod.gorax.dev` |
| `SSH_USER` | SSH username | `deploy` |
| `SSH_PORT` | SSH port (optional) | `22` |

### Secret Rotation Schedule

| Secret Type | Rotation Frequency | Responsibility |
|------------|-------------------|----------------|
| Database passwords | Every 90 days | DevOps team |
| API tokens | Every 180 days | Platform team |
| SSH keys | Every 365 days | Security team |
| Service account tokens | Every 90 days | DevOps team |

### Adding Secrets via CLI

```bash
# Add repository secret
gh secret set CODECOV_TOKEN --body "abc123def456..."

# Add secret from file
gh secret set SSH_PRIVATE_KEY < ~/.ssh/id_ed25519

# Add environment secret
gh secret set PRODUCTION_URL --env production --body "https://gorax.dev"

# List all secrets
gh secret list

# List environment secrets
gh api repos/{owner}/{repo}/environments/production/secrets | jq -r '.secrets[].name'
```

---

## Branch Protection Rules

### Main Branch Protection

Navigate to: **Settings → Branches → Add branch protection rule**

**Branch name pattern:** `main`

#### Protection Settings

**1. Require a pull request before merging**
- ✅ Enable
- **Required approvals:** 1
- ✅ Dismiss stale pull request approvals when new commits are pushed
- ✅ Require review from Code Owners (create `.github/CODEOWNERS` file)
- ✅ Require approval of the most recent reviewable push
- ❌ Require conversation resolution before merging (optional)

**2. Require status checks to pass before merging**
- ✅ Enable
- ✅ Require branches to be up to date before merging

**Required status checks:**
```
Go Tests
Go Lint
Frontend Tests
Frontend Lint
Coverage Threshold Check
Build Verification
Security Scanning / gosec
Security Scanning / npm-audit
CodeQL Analysis / analyze-go
CodeQL Analysis / analyze-typescript
Secrets Scanning / gitleaks
```

**How to identify status check names:**
1. Create a test PR
2. Wait for checks to run
3. Note the exact names from the checks list
4. Add each name to branch protection

**3. Require conversation resolution before merging**
- ⚠️ Optional - Enable if you want all PR comments resolved before merge

**4. Require signed commits**
- ✅ Enable (recommended for security)
- See [Signing Commits](#signing-commits) for setup

**5. Require linear history**
- ✅ Enable (prevents merge commits, enforces rebase/squash)

**6. Require deployments to succeed before merging**
- ❌ Disable (not applicable for main branch)

**7. Lock branch**
- ❌ Disable (would make branch read-only)

**8. Do not allow bypassing the above settings**
- ✅ Enable
- Consider adding specific teams/users who can bypass (e.g., security team for hotfixes)

**9. Restrict who can push to matching branches**
- ✅ Enable
- Add teams/users who can push directly (e.g., @gorax-team/platform-engineers)
- ⚠️ Use sparingly - normal workflow should use PRs

**10. Allow force pushes**
- ❌ Disable (prevents history rewriting)

**11. Allow deletions**
- ❌ Disable (prevents accidental branch deletion)

### Dev Branch Protection

Navigate to: **Settings → Branches → Add branch protection rule**

**Branch name pattern:** `dev`

Use similar settings as `main` with these differences:

| Setting | Main | Dev |
|---------|------|-----|
| Required approvals | 1 | 1 |
| Require linear history | ✅ | ❌ (optional) |
| Require signed commits | ✅ | ❌ (optional) |
| Require up-to-date branches | ✅ | ✅ |

### Feature Branch Naming Convention

Enforce branch naming via `.github/workflows/branch-naming.yml`:

```yaml
name: Branch Naming Convention

on:
  pull_request:
    types: [opened, synchronize, reopened, edited]

jobs:
  check-branch-name:
    runs-on: ubuntu-latest
    steps:
      - name: Check branch name
        run: |
          BRANCH="${{ github.head_ref }}"

          # Allow feature/, bugfix/, hotfix/, release/ prefixes
          if [[ ! $BRANCH =~ ^(feature|bugfix|hotfix|release|docs|test|refactor)/[A-Z]+-[0-9]+-[a-z0-9-]+$ ]]; then
            echo "❌ Invalid branch name: $BRANCH"
            echo "Branch name must follow pattern: <type>/<JIRA-123>-<description>"
            echo "Examples:"
            echo "  feature/RFLOW-123-add-new-action"
            echo "  bugfix/RFLOW-456-fix-memory-leak"
            exit 1
          fi

          echo "✅ Branch name is valid: $BRANCH"
```

---

## Repository Settings

### Code Security and Analysis

Navigate to: **Settings → Code security and analysis**

**Dependency graph:**
- ✅ Enable

**Dependabot:**
- ✅ Dependabot alerts
- ✅ Dependabot security updates
- ✅ Dependabot version updates (configure via `.github/dependabot.yml`)

**Code scanning:**
- ✅ CodeQL analysis (configured via `.github/workflows/codeql.yml`)
- ✅ Secret scanning
- ✅ Push protection (prevents committing secrets)

**Security policy:**
- ✅ Create SECURITY.md file with vulnerability reporting process

### Merge Button Settings

Navigate to: **Settings → General → Pull Requests**

Configure merge strategies:

- ✅ **Allow squash merging**
  - Default commit message: Pull request title and commit details
  - Keeps main/dev history clean
- ✅ **Allow merge commits**
  - For special cases where history preservation is needed
- ❌ **Allow rebase merging**
  - Disabled to prevent confusion and enforce squash merging
- ✅ **Automatically delete head branches**
  - Keeps repository clean after PR merge

### Webhooks (Optional)

Navigate to: **Settings → Webhooks → Add webhook**

Example integrations:

**Slack Notifications:**
- Payload URL: `https://hooks.slack.com/services/YOUR/WEBHOOK/URL`
- Content type: `application/json`
- Events: `Pull requests`, `Pushes`, `Releases`, `Workflow runs`

**External CI/CD:**
- Payload URL: Your CI/CD webhook endpoint
- Content type: `application/json`
- Secret: Generate with `openssl rand -hex 32`
- Events: `Push`, `Pull request`, `Release`

---

## Verification

### Automated Verification

Use the provided validation script:

```bash
# Run full validation
./scripts/validate-github-config.sh

# Check specific components
./scripts/validate-github-config.sh --environments
./scripts/validate-github-config.sh --secrets
./scripts/validate-github-config.sh --branch-protection
```

### Manual Verification Steps

#### 1. Verify Environments

```bash
# List environments
gh api repos/{owner}/{repo}/environments | jq -r '.environments[].name'

# Expected output:
# staging
# production

# Check staging environment
gh api repos/{owner}/{repo}/environments/staging | jq '.'

# Check production environment protection rules
gh api repos/{owner}/{repo}/environments/production | jq '.protection_rules'
```

#### 2. Verify Secrets

```bash
# List repository secrets
gh secret list

# List environment secrets
gh api repos/{owner}/{repo}/environments/production/secrets | jq -r '.secrets[].name'

# Expected secrets:
# CODECOV_TOKEN (repository)
# PRODUCTION_URL (production env)
# HEALTH_CHECK_TOKEN (production env)
# ... (others based on your infrastructure)
```

#### 3. Verify Branch Protection

```bash
# Check main branch protection
gh api repos/{owner}/{repo}/branches/main/protection | jq '.'

# Verify required status checks
gh api repos/{owner}/{repo}/branches/main/protection | jq '.required_status_checks.contexts'

# Check if force push is disabled
gh api repos/{owner}/{repo}/branches/main/protection | jq '.allow_force_pushes.enabled'
# Should output: false
```

#### 4. Test with Sample PR

Create a test pull request to verify all checks:

```bash
# Create test branch
git checkout -b test/verify-ci-config

# Make a trivial change
echo "# CI/CD Test" >> README.md
git add README.md
git commit -m "test: verify CI/CD configuration"
git push origin test/verify-ci-config

# Create PR
gh pr create \
  --title "Test: Verify CI/CD Configuration" \
  --body "Testing all status checks and branch protection rules" \
  --base dev

# Check PR status
PR_NUMBER=$(gh pr list --head test/verify-ci-config --json number --jq '.[0].number')
gh pr checks $PR_NUMBER

# Watch checks in real-time
gh pr checks $PR_NUMBER --watch

# Clean up after verification
gh pr close $PR_NUMBER --delete-branch
```

#### 5. Verify Deployment Workflows

Test staging deployment:

```bash
# Trigger staging deployment
gh workflow run deploy-staging.yml

# Monitor workflow
gh run list --workflow=deploy-staging.yml --limit 1
gh run watch
```

Test production deployment (requires approval):

```bash
# Trigger production deployment
gh workflow run deploy-production.yml

# Check pending approvals
gh api repos/{owner}/{repo}/actions/runs | jq '.workflow_runs[] | select(.status=="waiting")'

# Approve deployment (as a reviewer)
gh api repos/{owner}/{repo}/actions/runs/{run_id}/pending_deployments \
  -f environment_ids[]={env_id} \
  -f state=approved \
  -f comment="Approved for production deployment"
```

---

## Troubleshooting

### Common Issues and Solutions

#### Issue: "Resource not accessible by integration"

**Symptom:** GitHub Actions workflow fails with permissions error

**Solution:**
1. Navigate to **Settings → Actions → General**
2. Under **Workflow permissions**, select:
   - "Read and write permissions"
   - ✅ "Allow GitHub Actions to create and approve pull requests"

#### Issue: Environment not found in workflow

**Symptom:** Workflow fails with "Environment 'production' not found"

**Solution:**
1. Verify environment exists: `gh api repos/{owner}/{repo}/environments`
2. Check environment name in workflow matches exactly (case-sensitive)
3. Ensure workflow has correct environment configuration:
   ```yaml
   jobs:
     deploy:
       environment:
         name: production  # Must match exactly
         url: https://gorax.dev
   ```

#### Issue: Required status checks not appearing

**Symptom:** PR shows "Some checks haven't been completed yet"

**Solution:**
1. Check workflow is triggered for the branch:
   ```yaml
   on:
     pull_request:
       branches:
         - main
         - dev
   ```
2. Verify workflow has run at least once on the base branch
3. Check exact job names match in branch protection:
   ```bash
   # Get actual job names
   gh api repos/{owner}/{repo}/actions/runs?per_page=1 | jq '.workflow_runs[0].jobs_url'
   ```
4. Update branch protection with correct names

#### Issue: Cannot merge despite passing checks

**Symptom:** Merge button is disabled even with all checks passing

**Possible causes:**
1. Branch is not up to date with base branch
   - Solution: Update branch via "Update branch" button
2. Required reviews not provided
   - Solution: Request review from code owners
3. Conversations not resolved
   - Solution: Resolve all comments if "Require conversation resolution" is enabled
4. Missing required status check
   - Solution: Check which status check is missing and trigger it

#### Issue: Secrets not available in workflow

**Symptom:** Workflow shows empty or undefined secret values

**Solution:**
1. Verify secret is set in correct location:
   - Repository secrets: Available to all workflows
   - Environment secrets: Only available when using that environment
2. Check secret name matches exactly (case-sensitive):
   ```yaml
   # Correct
   ${{ secrets.PRODUCTION_URL }}
   # Wrong
   ${{ secrets.production_url }}
   ```
3. For environment secrets, ensure job specifies environment:
   ```yaml
   jobs:
     deploy:
       environment: production  # Required for env secrets
   ```

#### Issue: Deployment approval not triggering

**Symptom:** Production deployment doesn't wait for approval

**Solution:**
1. Verify environment protection rules:
   ```bash
   gh api repos/{owner}/{repo}/environments/production | jq '.protection_rules'
   ```
2. Check required reviewers are set:
   - Navigate to **Settings → Environments → production**
   - Ensure "Required reviewers" has at least one user/team
3. Verify workflow uses environment correctly:
   ```yaml
   jobs:
     deploy:
       environment:
         name: production  # Must specify environment
   ```

#### Issue: Branch protection prevents emergency hotfix

**Symptom:** Need to deploy critical fix but PR checks are failing

**Solution (Emergency Only):**
1. Option A: Use bypass permission (if configured)
   - Ask someone with bypass permission to merge
2. Option B: Temporarily disable protection
   - Navigate to **Settings → Branches**
   - Uncheck specific rules temporarily
   - ⚠️ Re-enable immediately after merge
3. Option C: Use workflow dispatch to deploy specific commit
   ```bash
   gh workflow run deploy-production.yml -f commit_sha=abc123
   ```

### Debug Commands

```bash
# Check GitHub CLI authentication
gh auth status

# Test API access
gh api user

# Get detailed branch protection info
gh api repos/{owner}/{repo}/branches/main/protection --paginate

# List all workflows
gh workflow list

# Get workflow run details
gh run view {run_id}

# Download workflow logs
gh run download {run_id}

# Check rate limit
gh api rate_limit
```

### Getting Help

**Internal Resources:**
- Slack: `#gorax-devops`
- Wiki: [CI/CD Troubleshooting](https://wiki.gorax.dev/cicd)
- On-call: PagerDuty @gorax-devops

**GitHub Resources:**
- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Branch Protection Rules](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/defining-the-mergeability-of-pull-requests/about-protected-branches)
- [GitHub Environments](https://docs.github.com/en/actions/deployment/targeting-different-environments/using-environments-for-deployment)
- [GitHub CLI Manual](https://cli.github.com/manual/)

---

## Maintenance

### Regular Review Schedule

| Task | Frequency | Owner |
|------|-----------|-------|
| Review and rotate secrets | Monthly | DevOps team |
| Audit branch protection rules | Quarterly | Security team |
| Review environment protection | Quarterly | Platform team |
| Update status check requirements | As needed | DevOps team |
| Review webhook configurations | Quarterly | Platform team |

### Quarterly Security Review Checklist

- [ ] Review all repository secrets
- [ ] Audit users with bypass permissions
- [ ] Verify 2FA enabled for all users with write access
- [ ] Review Dependabot alerts
- [ ] Check CodeQL findings
- [ ] Audit webhook configurations
- [ ] Review branch protection rules
- [ ] Check for leaked secrets in history
- [ ] Verify CODEOWNERS file is up to date

### Adding New Team Members

When adding new team members:

1. **Add to GitHub Organization/Repository**
   ```bash
   # Add user to repository
   gh api repos/{owner}/{repo}/collaborators/{username} -X PUT
   ```

2. **Configure Appropriate Role**
   - Read: Can view and clone
   - Triage: Can manage issues and PRs
   - Write: Can push to repository
   - Maintain: Can manage repository settings
   - Admin: Full access

3. **Add to Required Review Teams**
   - Update `.github/CODEOWNERS` if needed
   - Add to environment reviewers for production

4. **Provide Onboarding Documentation**
   - Share this configuration guide
   - Share `docs/DEVELOPER_GUIDE.md`
   - Share `docs/CI-CD.md`

### Deprecating Secrets

When rotating or deprecating secrets:

1. **Create new secret with temporary name**
   ```bash
   gh secret set NEW_API_TOKEN --body "new_value"
   ```

2. **Update workflows to use new secret**
   ```yaml
   - name: Use API token
     env:
       API_TOKEN: ${{ secrets.NEW_API_TOKEN }}
   ```

3. **Test thoroughly in staging**

4. **Rename secret to final name**
   ```bash
   gh secret delete OLD_API_TOKEN
   gh secret set API_TOKEN --body "new_value"
   ```

5. **Update workflows to final name**

6. **Verify in production**

---

## Signing Commits

### Setup GPG Signing

Required for protected branches with "Require signed commits" enabled.

**1. Generate GPG key:**

```bash
# Generate key
gpg --full-generate-key

# Choose:
# - RSA and RSA
# - 4096 bits
# - No expiration (or set expiration)
# - Your name and email (must match GitHub email)

# List keys
gpg --list-secret-keys --keyid-format=long

# Copy the key ID (after "sec rsa4096/")
```

**2. Configure Git:**

```bash
# Set signing key
git config --global user.signingkey YOUR_KEY_ID

# Enable commit signing by default
git config --global commit.gpgsign true

# Add to shell profile for GPG prompt
export GPG_TTY=$(tty)
```

**3. Add to GitHub:**

```bash
# Export public key
gpg --armor --export YOUR_KEY_ID

# Copy output and add to GitHub:
# Settings → SSH and GPG keys → New GPG key
```

**4. Verify:**

```bash
# Make a signed commit
git commit -S -m "test: verify GPG signing"

# Verify signature locally
git log --show-signature -1

# Push and verify on GitHub (should show "Verified" badge)
```

---

## Appendix

### Reference Files

- `.github/secrets.template.yml` - Template for required secrets
- `.github/environments/staging.json` - Staging environment configuration
- `.github/environments/production.json` - Production environment configuration
- `scripts/github-setup.sh` - Automated setup script
- `scripts/validate-github-config.sh` - Configuration validation script
- `terraform/github/` - Terraform configuration files

### Related Documentation

- [CI/CD Documentation](./CI-CD.md)
- [Deployment Guide](./DEPLOYMENT.md)
- [Security Guide](./SECURITY.md)
- [Developer Guide](./DEVELOPER_GUIDE.md)
- [Post-Deployment Checklist](./POST_DEPLOYMENT_CHECKLIST.md)

### Changelog

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0.0 | 2026-01-02 | @stherrien | Initial version |

---

**Document Version:** 1.0.0
**Last Updated:** 2026-01-02
**Maintained By:** DevOps Team (@gorax-team/devops)
