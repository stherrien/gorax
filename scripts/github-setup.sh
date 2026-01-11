#!/usr/bin/env bash

#######################################################################
# GitHub Repository Setup Script
#
# Automated setup of GitHub repository configuration including:
# - Environments (staging, production)
# - Secrets (repository and environment-specific)
# - Branch protection rules
# - Repository settings
#
# Usage:
#   ./scripts/github-setup.sh [OPTIONS]
#
# Options:
#   --dry-run           Show what would be done without making changes
#   --interactive       Prompt for values interactively
#   --skip-secrets      Skip secret configuration
#   --skip-protection   Skip branch protection setup
#   --help              Show this help message
#
# Requirements:
#   - GitHub CLI (gh) installed and authenticated
#   - Repository admin access
#   - jq for JSON processing
#######################################################################

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
DRY_RUN=false
INTERACTIVE=false
SKIP_SECRETS=false
SKIP_PROTECTION=false

# Repository information
REPO_OWNER="${GITHUB_OWNER:-stherrien}"
REPO_NAME="${GITHUB_REPO:-gorax}"
REPO_FULL="${REPO_OWNER}/${REPO_NAME}"

# URLs
STAGING_URL="${STAGING_URL:-https://staging.gorax.dev}"
PRODUCTION_URL="${PRODUCTION_URL:-https://gorax.dev}"

#######################################################################
# Helper Functions
#######################################################################

log_info() {
    echo -e "${BLUE}[INFO]${NC} $*"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $*"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $*"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $*" >&2
}

execute() {
    local description="$1"
    shift

    if [ "$DRY_RUN" = true ]; then
        log_info "[DRY RUN] Would execute: $description"
        log_info "Command: $*"
        return 0
    fi

    log_info "$description"
    if "$@"; then
        log_success "✓ $description"
        return 0
    else
        log_error "✗ Failed: $description"
        return 1
    fi
}

prompt() {
    local var_name="$1"
    local prompt_text="$2"
    local default_value="${3:-}"
    local is_secret="${4:-false}"

    if [ "$INTERACTIVE" = false ]; then
        return 0
    fi

    if [ "$is_secret" = true ]; then
        read -rsp "$prompt_text [$default_value]: " input
        echo
    else
        read -rp "$prompt_text [$default_value]: " input
    fi

    eval "$var_name=\"${input:-$default_value}\""
}

check_prerequisites() {
    log_info "Checking prerequisites..."

    # Check for gh CLI
    if ! command -v gh &> /dev/null; then
        log_error "GitHub CLI (gh) is not installed"
        log_info "Install with: brew install gh"
        exit 1
    fi

    # Check for jq
    if ! command -v jq &> /dev/null; then
        log_error "jq is not installed"
        log_info "Install with: brew install jq"
        exit 1
    fi

    # Check gh authentication
    if ! gh auth status &> /dev/null; then
        log_error "Not authenticated with GitHub CLI"
        log_info "Run: gh auth login"
        exit 1
    fi

    # Check repository access
    if ! gh repo view "$REPO_FULL" &> /dev/null; then
        log_error "Cannot access repository: $REPO_FULL"
        log_info "Ensure you have admin access to the repository"
        exit 1
    fi

    log_success "All prerequisites met"
}

#######################################################################
# Environment Setup
#######################################################################

setup_staging_environment() {
    log_info "Setting up staging environment..."

    # Check if environment exists
    if gh api "repos/$REPO_FULL/environments/staging" &> /dev/null; then
        log_warning "Staging environment already exists"
        return 0
    fi

    # Create staging environment
    execute "Creating staging environment" \
        gh api "repos/$REPO_FULL/environments/staging" -X PUT \
        -f "deployment_branch_policy[protected_branches]=false" \
        -f "deployment_branch_policy[custom_branch_policies]=true"

    # Add deployment branch policy for dev
    execute "Adding dev branch to staging deployment policy" \
        gh api "repos/$REPO_FULL/environments/staging/deployment-branch-policies" -X POST \
        -f "name=dev" \
        -f "type=branch"

    # Add environment URL
    execute "Setting staging environment URL" \
        gh api "repos/$REPO_FULL/environments/staging" -X PUT \
        -f "url=$STAGING_URL"
}

setup_production_environment() {
    log_info "Setting up production environment..."

    # Prompt for reviewer user ID if interactive
    local reviewer_id="${PRODUCTION_REVIEWER_ID:-}"
    if [ "$INTERACTIVE" = true ] && [ -z "$reviewer_id" ]; then
        echo
        log_info "To get your user ID, run: gh api users/{username} | jq .id"
        read -rp "Enter GitHub user ID for production approvals: " reviewer_id
    fi

    # Check if environment exists
    if gh api "repos/$REPO_FULL/environments/production" &> /dev/null; then
        log_warning "Production environment already exists"
    else
        # Create production environment
        execute "Creating production environment" \
            gh api "repos/$REPO_FULL/environments/production" -X PUT \
            -f "deployment_branch_policy[protected_branches]=false" \
            -f "deployment_branch_policy[custom_branch_policies]=true"
    fi

    # Add deployment branch policy for main
    execute "Adding main branch to production deployment policy" \
        gh api "repos/$REPO_FULL/environments/production/deployment-branch-policies" -X POST \
        -f "name=main" \
        -f "type=branch" 2> /dev/null || log_warning "Branch policy may already exist"

    # Add required reviewers if provided
    if [ -n "$reviewer_id" ]; then
        execute "Adding required reviewer to production" \
            gh api "repos/$REPO_FULL/environments/production" -X PUT \
            -f "reviewers[0][type]=User" \
            -f "reviewers[0][id]=$reviewer_id" \
            -f "deployment_branch_policy[protected_branches]=false" \
            -f "deployment_branch_policy[custom_branch_policies]=true"
    else
        log_warning "No reviewer specified for production environment"
        log_info "You can add reviewers manually in GitHub Settings → Environments → production"
    fi

    # Add wait timer (5 minutes)
    execute "Setting wait timer for production" \
        gh api "repos/$REPO_FULL/environments/production" -X PUT \
        -f "wait_timer=5" \
        -f "deployment_branch_policy[protected_branches]=false" \
        -f "deployment_branch_policy[custom_branch_policies]=true"

    # Add environment URL
    execute "Setting production environment URL" \
        gh api "repos/$REPO_FULL/environments/production" -X PUT \
        -f "url=$PRODUCTION_URL"
}

#######################################################################
# Secrets Setup
#######################################################################

setup_repository_secrets() {
    if [ "$SKIP_SECRETS" = true ]; then
        log_info "Skipping repository secrets setup"
        return 0
    fi

    log_info "Setting up repository secrets..."

    # Codecov token (optional)
    local codecov_token="${CODECOV_TOKEN:-}"
    if [ "$INTERACTIVE" = true ]; then
        prompt codecov_token "Codecov token (leave empty to skip)" "" true
    fi

    if [ -n "$codecov_token" ]; then
        execute "Setting CODECOV_TOKEN" \
            gh secret set CODECOV_TOKEN --body "$codecov_token" --repo "$REPO_FULL"
    else
        log_info "Skipping CODECOV_TOKEN (not provided)"
    fi
}

setup_staging_secrets() {
    if [ "$SKIP_SECRETS" = true ]; then
        log_info "Skipping staging secrets setup"
        return 0
    fi

    log_info "Setting up staging environment secrets..."

    # Staging URL
    execute "Setting STAGING_URL" \
        gh secret set STAGING_URL --env staging --body "$STAGING_URL" --repo "$REPO_FULL"

    # Staging DB password
    local staging_db_password="${STAGING_DB_PASSWORD:-}"
    if [ "$INTERACTIVE" = true ]; then
        prompt staging_db_password "Staging database password" "" true
    fi

    if [ -n "$staging_db_password" ]; then
        execute "Setting STAGING_DB_PASSWORD" \
            gh secret set STAGING_DB_PASSWORD --env staging --body "$staging_db_password" --repo "$REPO_FULL"
    else
        log_warning "STAGING_DB_PASSWORD not set - configure manually later"
    fi

    # Staging Redis password
    local staging_redis_password="${STAGING_REDIS_PASSWORD:-}"
    if [ "$INTERACTIVE" = true ]; then
        prompt staging_redis_password "Staging Redis password" "" true
    fi

    if [ -n "$staging_redis_password" ]; then
        execute "Setting STAGING_REDIS_PASSWORD" \
            gh secret set STAGING_REDIS_PASSWORD --env staging --body "$staging_redis_password" --repo "$REPO_FULL"
    else
        log_info "STAGING_REDIS_PASSWORD not set - configure manually if needed"
    fi
}

setup_production_secrets() {
    if [ "$SKIP_SECRETS" = true ]; then
        log_info "Skipping production secrets setup"
        return 0
    fi

    log_info "Setting up production environment secrets..."

    # Production URL
    execute "Setting PRODUCTION_URL" \
        gh secret set PRODUCTION_URL --env production --body "$PRODUCTION_URL" --repo "$REPO_FULL"

    # Production DB password
    local production_db_password="${PRODUCTION_DB_PASSWORD:-}"
    if [ "$INTERACTIVE" = true ]; then
        prompt production_db_password "Production database password" "" true
    fi

    if [ -n "$production_db_password" ]; then
        execute "Setting PRODUCTION_DB_PASSWORD" \
            gh secret set PRODUCTION_DB_PASSWORD --env production --body "$production_db_password" --repo "$REPO_FULL"
    else
        log_warning "PRODUCTION_DB_PASSWORD not set - MUST configure manually"
    fi

    # Production Redis password
    local production_redis_password="${PRODUCTION_REDIS_PASSWORD:-}"
    if [ "$INTERACTIVE" = true ]; then
        prompt production_redis_password "Production Redis password" "" true
    fi

    if [ -n "$production_redis_password" ]; then
        execute "Setting PRODUCTION_REDIS_PASSWORD" \
            gh secret set PRODUCTION_REDIS_PASSWORD --env production --body "$production_redis_password" --repo "$REPO_FULL"
    else
        log_warning "PRODUCTION_REDIS_PASSWORD not set - configure manually if needed"
    fi

    # Health check token
    local health_check_token="${HEALTH_CHECK_TOKEN:-}"
    if [ "$INTERACTIVE" = true ]; then
        prompt health_check_token "Health check token" "" true
    fi

    if [ -n "$health_check_token" ]; then
        execute "Setting HEALTH_CHECK_TOKEN" \
            gh secret set HEALTH_CHECK_TOKEN --env production --body "$health_check_token" --repo "$REPO_FULL"
    else
        log_warning "HEALTH_CHECK_TOKEN not set - configure manually"
    fi
}

#######################################################################
# Branch Protection Setup
#######################################################################

setup_main_branch_protection() {
    if [ "$SKIP_PROTECTION" = true ]; then
        log_info "Skipping main branch protection setup"
        return 0
    fi

    log_info "Setting up main branch protection..."

    # Note: GitHub CLI doesn't support all branch protection options
    # Some settings may need to be configured via web UI

    execute "Enabling main branch protection" \
        gh api "repos/$REPO_FULL/branches/main/protection" -X PUT \
        --input - <<'EOF'
{
  "required_status_checks": {
    "strict": true,
    "contexts": [
      "Go Tests",
      "Go Lint",
      "Frontend Tests",
      "Frontend Lint",
      "Coverage Threshold Check",
      "Build Verification",
      "Security Scanning / gosec",
      "Security Scanning / npm-audit",
      "CodeQL Analysis / analyze-go",
      "CodeQL Analysis / analyze-typescript",
      "Secrets Scanning / gitleaks"
    ]
  },
  "enforce_admins": true,
  "required_pull_request_reviews": {
    "dismissal_restrictions": {},
    "dismiss_stale_reviews": true,
    "require_code_owner_reviews": true,
    "required_approving_review_count": 1,
    "require_last_push_approval": true
  },
  "restrictions": null,
  "required_linear_history": true,
  "allow_force_pushes": false,
  "allow_deletions": false,
  "block_creations": false,
  "required_conversation_resolution": false,
  "lock_branch": false,
  "allow_fork_syncing": true
}
EOF

    log_success "Main branch protection configured"
    log_warning "Note: Some settings (signed commits) must be enabled via web UI"
}

setup_dev_branch_protection() {
    if [ "$SKIP_PROTECTION" = true ]; then
        log_info "Skipping dev branch protection setup"
        return 0
    fi

    log_info "Setting up dev branch protection..."

    execute "Enabling dev branch protection" \
        gh api "repos/$REPO_FULL/branches/dev/protection" -X PUT \
        --input - <<'EOF'
{
  "required_status_checks": {
    "strict": true,
    "contexts": [
      "Go Tests",
      "Go Lint",
      "Frontend Tests",
      "Frontend Lint",
      "Coverage Threshold Check",
      "Build Verification"
    ]
  },
  "enforce_admins": false,
  "required_pull_request_reviews": {
    "dismissal_restrictions": {},
    "dismiss_stale_reviews": true,
    "require_code_owner_reviews": false,
    "required_approving_review_count": 1,
    "require_last_push_approval": false
  },
  "restrictions": null,
  "required_linear_history": false,
  "allow_force_pushes": false,
  "allow_deletions": false,
  "block_creations": false,
  "required_conversation_resolution": false,
  "lock_branch": false,
  "allow_fork_syncing": true
}
EOF

    log_success "Dev branch protection configured"
}

#######################################################################
# Repository Settings
#######################################################################

update_repository_settings() {
    log_info "Updating repository settings..."

    execute "Enabling automatic branch deletion" \
        gh api "repos/$REPO_FULL" -X PATCH \
        -f "delete_branch_on_merge=true"

    execute "Configuring merge strategies" \
        gh api "repos/$REPO_FULL" -X PATCH \
        -f "allow_squash_merge=true" \
        -f "allow_merge_commit=true" \
        -f "allow_rebase_merge=false" \
        -f "allow_auto_merge=false"

    execute "Enabling vulnerability alerts" \
        gh api "repos/$REPO_FULL/vulnerability-alerts" -X PUT

    log_success "Repository settings updated"
}

#######################################################################
# Verification
#######################################################################

verify_setup() {
    log_info "Verifying configuration..."

    echo
    log_info "=== Environments ==="
    if gh api "repos/$REPO_FULL/environments" | jq -e '.environments[] | select(.name=="staging")' &> /dev/null; then
        log_success "✓ Staging environment exists"
    else
        log_error "✗ Staging environment not found"
    fi

    if gh api "repos/$REPO_FULL/environments" | jq -e '.environments[] | select(.name=="production")' &> /dev/null; then
        log_success "✓ Production environment exists"
    else
        log_error "✗ Production environment not found"
    fi

    echo
    log_info "=== Branch Protection ==="
    if gh api "repos/$REPO_FULL/branches/main/protection" &> /dev/null; then
        log_success "✓ Main branch is protected"
    else
        log_error "✗ Main branch is not protected"
    fi

    if gh api "repos/$REPO_FULL/branches/dev/protection" &> /dev/null; then
        log_success "✓ Dev branch is protected"
    else
        log_error "✗ Dev branch is not protected"
    fi

    echo
    log_info "=== Secrets ==="
    local secret_count
    secret_count=$(gh secret list --repo "$REPO_FULL" | wc -l)
    log_info "Repository secrets configured: $secret_count"

    echo
    log_info "=== Next Steps ==="
    log_info "1. Review configuration in GitHub UI:"
    log_info "   https://github.com/$REPO_FULL/settings"
    log_info "2. Add any missing secrets via GitHub UI or gh CLI"
    log_info "3. Configure signed commits if required"
    log_info "4. Run validation script: ./scripts/validate-github-config.sh"
}

#######################################################################
# Main
#######################################################################

show_help() {
    cat << EOF
GitHub Repository Setup Script

Automated setup of GitHub repository configuration including environments,
secrets, branch protection, and repository settings.

Usage:
  $0 [OPTIONS]

Options:
  --dry-run           Show what would be done without making changes
  --interactive       Prompt for values interactively
  --skip-secrets      Skip secret configuration
  --skip-protection   Skip branch protection setup
  --help              Show this help message

Environment Variables:
  GITHUB_OWNER              Repository owner (default: stherrien)
  GITHUB_REPO               Repository name (default: gorax)
  STAGING_URL               Staging environment URL
  PRODUCTION_URL            Production environment URL
  PRODUCTION_REVIEWER_ID    GitHub user ID for production approvals
  CODECOV_TOKEN             Codecov upload token
  STAGING_DB_PASSWORD       Staging database password
  STAGING_REDIS_PASSWORD    Staging Redis password
  PRODUCTION_DB_PASSWORD    Production database password
  PRODUCTION_REDIS_PASSWORD Production Redis password
  HEALTH_CHECK_TOKEN        Health check token

Examples:
  # Dry run to see what would happen
  $0 --dry-run

  # Interactive setup with prompts
  $0 --interactive

  # Skip secrets configuration
  $0 --skip-secrets

  # Configure with environment variables
  PRODUCTION_REVIEWER_ID=12345678 $0

Requirements:
  - GitHub CLI (gh) installed and authenticated
  - Repository admin access
  - jq for JSON processing

For more information, see:
  docs/GITHUB_CONFIGURATION_GUIDE.md

EOF
}

main() {
    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --dry-run)
                DRY_RUN=true
                shift
                ;;
            --interactive)
                INTERACTIVE=true
                shift
                ;;
            --skip-secrets)
                SKIP_SECRETS=true
                shift
                ;;
            --skip-protection)
                SKIP_PROTECTION=true
                shift
                ;;
            --help)
                show_help
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done

    # Print banner
    echo
    log_info "=========================================="
    log_info "  GitHub Repository Setup"
    log_info "=========================================="
    log_info "Repository: $REPO_FULL"
    [ "$DRY_RUN" = true ] && log_warning "DRY RUN MODE - No changes will be made"
    echo

    # Check prerequisites
    check_prerequisites

    # Setup environments
    echo
    log_info "=== Setting up environments ==="
    setup_staging_environment
    setup_production_environment

    # Setup secrets
    echo
    log_info "=== Setting up secrets ==="
    setup_repository_secrets
    setup_staging_secrets
    setup_production_secrets

    # Setup branch protection
    echo
    log_info "=== Setting up branch protection ==="
    setup_main_branch_protection
    setup_dev_branch_protection

    # Update repository settings
    echo
    log_info "=== Updating repository settings ==="
    update_repository_settings

    # Verify setup
    if [ "$DRY_RUN" = false ]; then
        echo
        verify_setup
    fi

    # Done
    echo
    log_success "=========================================="
    log_success "  GitHub setup completed!"
    log_success "=========================================="

    if [ "$DRY_RUN" = true ]; then
        log_info "Run without --dry-run to apply changes"
    fi
}

# Run main function
main "$@"
