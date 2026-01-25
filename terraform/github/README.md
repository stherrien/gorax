# GitHub Repository Configuration with Terraform

Terraform configuration for automating GitHub repository settings including environments, secrets, branch protection, and CI/CD integration.

## Prerequisites

1. **Terraform** installed (>= 1.0.0)
   ```bash
   brew install terraform
   ```

2. **GitHub Personal Access Token** with permissions:
   - `repo` (full control)
   - `admin:org` (if using organization)
   - `delete_repo` (optional, for cleanup)

3. **GitHub CLI** (optional, for helper commands)
   ```bash
   brew install gh
   gh auth login
   ```

## Quick Start

### 1. Clone Repository

```bash
git clone https://github.com/stherrien/gorax.git
cd gorax/terraform/github
```

### 2. Configure Variables

```bash
# Copy example configuration
cp terraform.tfvars.example terraform.tfvars

# Edit with your values
vim terraform.tfvars
```

**Required variables:**
- `github_owner`: Your GitHub username or organization
- `github_token`: Personal access token
- `production_reviewers`: List of user IDs who can approve production deployments

**Get GitHub user ID:**
```bash
gh api users/{username} | jq .id
```

### 3. Initialize Terraform

```bash
terraform init
```

### 4. Plan Changes

```bash
# Review what will be created/changed
terraform plan
```

### 5. Apply Configuration

```bash
# Apply with confirmation
terraform apply

# Or auto-approve (use with caution)
terraform apply -auto-approve
```

## What Gets Configured

### Repository Settings
- Description, homepage, topics
- Features (issues, projects, discussions)
- Security settings (Dependabot, secret scanning)
- Merge strategies

### Environments
- **staging**: Auto-deploy from `dev` branch
- **production**: Requires approval, deploys from `main` branch
- **development** (optional): For local testing

### Branch Protection
- **main**: Strict protection, requires reviews and all checks
- **dev**: Standard protection, allows more flexibility
- **release/\***: (optional) Release branch protection
- **hotfix/\***: (optional) Emergency fix protection

### Secrets (Optional)
- Repository-level secrets
- Environment-specific secrets
- Infrastructure secrets (Kubernetes, AWS ECS, or SSH)

**Note:** Managing secrets via Terraform stores them in state file. Consider using GitHub UI or `gh` CLI for sensitive secrets.

### Labels
- Priority labels (critical, high, medium, low)
- Type labels (bug, feature, enhancement, documentation, security)

## Configuration Options

### Secret Management

By default, Terraform does not manage secrets (`manage_secrets = false`). To enable:

```hcl
# terraform.tfvars
manage_secrets = true

# Provide secret values
codecov_token = "abc123..."
production_db_password = "strong_password"
```

**Security Warning:** Secrets will be stored in Terraform state. Use one of these alternatives:
1. Set secrets via GitHub UI
2. Use `gh` CLI: `gh secret set SECRET_NAME`
3. Use external secret management (Vault, AWS Secrets Manager)

### Deployment Type

Choose your deployment infrastructure:

```hcl
# Kubernetes
deployment_type = "kubernetes"
kubeconfig      = "base64_encoded_config"
k8s_cluster_url = "https://k8s.example.com"
k8s_token       = "token"
k8s_namespace   = "gorax"

# AWS ECS
deployment_type       = "aws-ecs"
aws_access_key_id     = "AKIA..."
aws_secret_access_key = "..."
aws_region            = "us-east-1"
ecs_cluster           = "gorax-production"
ecs_service           = "gorax-api"

# SSH
deployment_type = "ssh"
ssh_private_key = "-----BEGIN OPENSSH PRIVATE KEY-----..."
ssh_host        = "prod.gorax.dev"
ssh_user        = "deploy"
```

### Branch Protection Customization

```hcl
# Require signed commits
require_signed_commits = true

# Require conversation resolution
require_conversation_resolution = true

# Allow specific users/teams to push directly
main_branch_push_allowances = [
  "/users/12345678",  # Specific user
  "/teams/87654321",  # Specific team
]

# Enable additional branch patterns
enable_release_branch_protection = true
enable_hotfix_branch_protection  = true
```

## Usage Examples

### Full Configuration

```bash
# Initialize
terraform init

# Review planned changes
terraform plan -out=tfplan

# Apply changes
terraform apply tfplan

# View outputs
terraform output
```

### Update Specific Resource

```bash
# Update only branch protection
terraform apply -target=github_branch_protection.main

# Update only environments
terraform apply -target=github_repository_environment.production
```

### Import Existing Configuration

If you have existing GitHub configuration:

```bash
# Import repository
terraform import github_repository.main gorax

# Import branch protection
terraform import github_branch_protection.main stherrien/gorax:main

# Import environment
terraform import github_repository_environment.production stherrien/gorax:production
```

### Destroy Resources

```bash
# Preview what will be destroyed
terraform plan -destroy

# Destroy (use with extreme caution)
terraform destroy
```

## State Management

### Local State (Default)

State is stored locally in `terraform.tfstate`. **Do not commit this file** (it contains secrets).

```bash
# .gitignore already includes
terraform.tfstate
terraform.tfstate.backup
*.tfvars
```

### Remote State (Recommended for Teams)

Configure remote state backend:

```hcl
# main.tf
terraform {
  backend "s3" {
    bucket         = "gorax-terraform-state"
    key            = "github/terraform.tfstate"
    region         = "us-east-1"
    encrypt        = true
    dynamodb_table = "gorax-terraform-locks"
  }
}
```

Initialize with backend:

```bash
terraform init -backend-config="bucket=your-bucket"
```

## Validation

### Verify Configuration

```bash
# Check Terraform syntax
terraform validate

# Format code
terraform fmt -recursive

# Lint with tflint
tflint

# Security scan with tfsec
tfsec .
```

### Verify GitHub Configuration

After applying, verify with GitHub CLI:

```bash
# List environments
gh api repos/stherrien/gorax/environments | jq -r '.environments[].name'

# Check branch protection
gh api repos/stherrien/gorax/branches/main/protection | jq .

# List secrets (names only)
gh secret list

# View environment secrets
gh api repos/stherrien/gorax/environments/production/secrets | jq -r '.secrets[].name'
```

## Troubleshooting

### Error: Resource Already Exists

If Terraform tries to create resources that already exist:

```bash
# Import existing resource
terraform import github_repository.main stherrien/gorax
terraform import github_branch_protection.main stherrien/gorax:main
```

### Error: Insufficient Permissions

Ensure your GitHub token has required permissions:
- `repo` (full control of private repositories)
- `admin:org` (full control of orgs and teams, read and write org projects)

```bash
# Check token permissions
gh auth status

# Re-authenticate with correct scopes
gh auth login --scopes repo,admin:org
```

### Error: Status Check Context Not Found

If required status checks don't exist:

1. Create a test PR to trigger workflows
2. Note exact status check names
3. Update `branch-protection.tf`:
   ```hcl
   contexts = [
     "Exact Status Check Name",
     # ...
   ]
   ```

### State File Issues

If state becomes corrupted:

```bash
# Backup current state
cp terraform.tfstate terraform.tfstate.backup

# Pull latest state (if using remote backend)
terraform state pull

# List resources in state
terraform state list

# Remove problematic resource
terraform state rm github_repository.main

# Re-import
terraform import github_repository.main stherrien/gorax
```

## Maintenance

### Regular Updates

```bash
# Update provider versions
terraform init -upgrade

# Review and apply changes
terraform plan
terraform apply
```

### Drift Detection

```bash
# Check for configuration drift
terraform plan -detailed-exitcode

# Exit codes:
# 0 = no changes
# 1 = error
# 2 = changes detected
```

### Rotate Secrets

```bash
# Update secret in terraform.tfvars
vim terraform.tfvars

# Apply changes
terraform apply -target=github_actions_secret.production_db_password
```

## Best Practices

1. **Never commit `terraform.tfvars`** - Contains sensitive data
2. **Use remote state** - For team collaboration
3. **Enable state locking** - Prevent concurrent modifications
4. **Review plans carefully** - Before applying
5. **Use version control** - For Terraform code
6. **Tag releases** - Of Terraform configurations
7. **Document changes** - In commit messages
8. **Test in staging** - Before production
9. **Backup state** - Regularly
10. **Rotate credentials** - Periodically

## CI/CD Integration

Automate Terraform runs in GitHub Actions:

```yaml
# .github/workflows/terraform.yml
name: Terraform

on:
  push:
    paths:
      - 'terraform/github/**'
  pull_request:
    paths:
      - 'terraform/github/**'

jobs:
  terraform:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: hashicorp/setup-terraform@v3
      - name: Terraform Init
        run: terraform init
        working-directory: terraform/github
      - name: Terraform Plan
        run: terraform plan
        working-directory: terraform/github
        env:
          GITHUB_TOKEN: ${{ secrets.GH_PAT }}
```

## Related Documentation

- [GitHub Configuration Guide](../../docs/GITHUB_CONFIGURATION_GUIDE.md)
- [Security Checklist](../../docs/GITHUB_SECURITY_CHECKLIST.md)
- [Deployment Guide](../../docs/DEPLOYMENT.md)
- [CI/CD Documentation](../../docs/CI-CD.md)

## Support

For issues or questions:
- Open an issue: https://github.com/stherrien/gorax/issues
- Slack: `#gorax-devops`
- Documentation: https://docs.gorax.dev

---

**Version:** 1.0.0
**Last Updated:** 2026-01-02
**Maintained By:** DevOps Team
