#!/usr/bin/env bash

#######################################################################
# GitHub Configuration Validation Script
#
# Validates GitHub repository configuration including:
# - Environments (staging, production)
# - Secrets (repository and environment-specific)
# - Branch protection rules
# - Repository settings
#
# Usage:
#   ./scripts/validate-github-config.sh [OPTIONS]
#
# Options:
#   --environments      Check only environments
#   --secrets          Check only secrets
#   --branch-protection Check only branch protection
#   --all              Check everything (default)
#   --verbose          Show detailed output
#   --json             Output results as JSON
#   --help             Show this help message
#
# Exit codes:
#   0 - All checks passed
#   1 - Some checks failed
#   2 - Critical error (cannot connect, etc.)
#
# Requirements:
#   - GitHub CLI (gh) installed and authenticated
#   - jq for JSON processing
#######################################################################

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Repository information
REPO_OWNER="${GITHUB_OWNER:-stherrien}"
REPO_NAME="${GITHUB_REPO:-gorax}"
REPO_FULL="${REPO_OWNER}/${REPO_NAME}"

# Check options
CHECK_ENVIRONMENTS=true
CHECK_SECRETS=true
CHECK_BRANCH_PROTECTION=true
VERBOSE=false
JSON_OUTPUT=false

# Results tracking
TOTAL_CHECKS=0
PASSED_CHECKS=0
FAILED_CHECKS=0
WARNING_CHECKS=0

declare -a FAILURES=()
declare -a WARNINGS=()

#######################################################################
# Helper Functions
#######################################################################

log_info() {
    if [ "$JSON_OUTPUT" = false ]; then
        echo -e "${BLUE}[INFO]${NC} $*"
    fi
}

log_success() {
    if [ "$JSON_OUTPUT" = false ]; then
        echo -e "${GREEN}[✓]${NC} $*"
    fi
}

log_warning() {
    if [ "$JSON_OUTPUT" = false ]; then
        echo -e "${YELLOW}[⚠]${NC} $*"
    fi
}

log_error() {
    if [ "$JSON_OUTPUT" = false ]; then
        echo -e "${RED}[✗]${NC} $*" >&2
    fi
}

log_verbose() {
    if [ "$VERBOSE" = true ] && [ "$JSON_OUTPUT" = false ]; then
        echo -e "${CYAN}[DEBUG]${NC} $*"
    fi
}

check() {
    local description="$1"
    shift

    TOTAL_CHECKS=$((TOTAL_CHECKS + 1))
    log_verbose "Checking: $description"

    if "$@" &> /dev/null; then
        PASSED_CHECKS=$((PASSED_CHECKS + 1))
        log_success "$description"
        return 0
    else
        FAILED_CHECKS=$((FAILED_CHECKS + 1))
        log_error "$description"
        FAILURES+=("$description")
        return 1
    fi
}

warn() {
    local description="$1"
    shift

    TOTAL_CHECKS=$((TOTAL_CHECKS + 1))
    log_verbose "Checking (warning): $description"

    if "$@" &> /dev/null; then
        PASSED_CHECKS=$((PASSED_CHECKS + 1))
        log_success "$description"
        return 0
    else
        WARNING_CHECKS=$((WARNING_CHECKS + 1))
        log_warning "$description (optional)"
        WARNINGS+=("$description")
        return 1
    fi
}

check_prerequisites() {
    log_info "Checking prerequisites..."

    # Check for gh CLI
    if ! command -v gh &> /dev/null; then
        log_error "GitHub CLI (gh) is not installed"
        log_info "Install with: brew install gh"
        exit 2
    fi

    # Check for jq
    if ! command -v jq &> /dev/null; then
        log_error "jq is not installed"
        log_info "Install with: brew install jq"
        exit 2
    fi

    # Check gh authentication
    if ! gh auth status &> /dev/null; then
        log_error "Not authenticated with GitHub CLI"
        log_info "Run: gh auth login"
        exit 2
    fi

    # Check repository access
    if ! gh repo view "$REPO_FULL" &> /dev/null; then
        log_error "Cannot access repository: $REPO_FULL"
        log_info "Ensure you have access to the repository"
        exit 2
    fi

    log_success "Prerequisites OK"
    echo
}

#######################################################################
# Environment Validation
#######################################################################

validate_environments() {
    if [ "$CHECK_ENVIRONMENTS" = false ]; then
        return 0
    fi

    log_info "Validating environments..."

    # Check staging environment exists
    check "Staging environment exists" \
        gh api "repos/$REPO_FULL/environments/staging"

    # Check staging deployment branch policy
    if gh api "repos/$REPO_FULL/environments/staging/deployment-branch-policies" 2>/dev/null | jq -e '.[] | select(.name=="dev")' > /dev/null; then
        check "Staging allows deployment from dev branch" true
    else
        check "Staging allows deployment from dev branch" false
    fi

    # Check production environment exists
    check "Production environment exists" \
        gh api "repos/$REPO_FULL/environments/production"

    # Check production has required reviewers
    local reviewer_count
    reviewer_count=$(gh api "repos/$REPO_FULL/environments/production" | jq '.reviewers | length' 2>/dev/null || echo "0")
    if [ "$reviewer_count" -gt 0 ]; then
        check "Production has required reviewers" true
        log_verbose "  Found $reviewer_count reviewer(s)"
    else
        check "Production has required reviewers" false
    fi

    # Check production wait timer
    local wait_timer
    wait_timer=$(gh api "repos/$REPO_FULL/environments/production" | jq -r '.wait_timer' 2>/dev/null || echo "null")
    if [ "$wait_timer" != "null" ] && [ "$wait_timer" -gt 0 ]; then
        check "Production has wait timer configured" true
        log_verbose "  Wait timer: ${wait_timer} minutes"
    else
        warn "Production has wait timer configured" false
    fi

    # Check production deployment branch policy
    if gh api "repos/$REPO_FULL/environments/production/deployment-branch-policies" 2>/dev/null | jq -e '.[] | select(.name=="main")' > /dev/null; then
        check "Production allows deployment from main branch only" true
    else
        check "Production allows deployment from main branch only" false
    fi

    echo
}

#######################################################################
# Secrets Validation
#######################################################################

validate_secrets() {
    if [ "$CHECK_SECRETS" = false ]; then
        return 0
    fi

    log_info "Validating secrets..."

    # Repository secrets
    log_verbose "Checking repository secrets..."

    # CODECOV_TOKEN is optional
    if gh secret list --repo "$REPO_FULL" | grep -q "CODECOV_TOKEN"; then
        log_success "CODECOV_TOKEN is configured (optional)"
    else
        log_verbose "  CODECOV_TOKEN not found (optional)"
    fi

    # Staging environment secrets
    log_verbose "Checking staging environment secrets..."

    check "STAGING_URL secret exists" \
        gh api "repos/$REPO_FULL/environments/staging/secrets" | jq -e '.secrets[] | select(.name=="STAGING_URL")'

    warn "STAGING_DB_PASSWORD secret exists" \
        gh api "repos/$REPO_FULL/environments/staging/secrets" | jq -e '.secrets[] | select(.name=="STAGING_DB_PASSWORD")'

    # Production environment secrets
    log_verbose "Checking production environment secrets..."

    check "PRODUCTION_URL secret exists" \
        gh api "repos/$REPO_FULL/environments/production/secrets" | jq -e '.secrets[] | select(.name=="PRODUCTION_URL")'

    warn "PRODUCTION_DB_PASSWORD secret exists" \
        gh api "repos/$REPO_FULL/environments/production/secrets" | jq -e '.secrets[] | select(.name=="PRODUCTION_DB_PASSWORD")'

    warn "HEALTH_CHECK_TOKEN secret exists" \
        gh api "repos/$REPO_FULL/environments/production/secrets" | jq -e '.secrets[] | select(.name=="HEALTH_CHECK_TOKEN")'

    # Check for infrastructure-specific secrets
    log_verbose "Checking infrastructure secrets..."

    local has_k8s=false
    local has_aws=false
    local has_ssh=false

    if gh secret list --repo "$REPO_FULL" | grep -q "KUBECONFIG"; then
        has_k8s=true
        log_verbose "  Kubernetes deployment detected"
    fi

    if gh secret list --repo "$REPO_FULL" | grep -q "AWS_ACCESS_KEY_ID"; then
        has_aws=true
        log_verbose "  AWS ECS deployment detected"
    fi

    if gh secret list --repo "$REPO_FULL" | grep -q "SSH_PRIVATE_KEY"; then
        has_ssh=true
        log_verbose "  SSH deployment detected"
    fi

    if [ "$has_k8s" = false ] && [ "$has_aws" = false ] && [ "$has_ssh" = false ]; then
        log_warning "No deployment infrastructure secrets found"
        log_info "  Configure one of: Kubernetes, AWS ECS, or SSH"
    fi

    echo
}

#######################################################################
# Branch Protection Validation
#######################################################################

validate_branch_protection() {
    if [ "$CHECK_BRANCH_PROTECTION" = false ]; then
        return 0
    fi

    log_info "Validating branch protection..."

    # Main branch protection
    log_verbose "Checking main branch protection..."

    check "Main branch is protected" \
        gh api "repos/$REPO_FULL/branches/main/protection"

    check "Main branch requires pull request reviews" \
        gh api "repos/$REPO_FULL/branches/main/protection" | jq -e '.required_pull_request_reviews'

    local required_approvals
    required_approvals=$(gh api "repos/$REPO_FULL/branches/main/protection" 2>/dev/null | jq -r '.required_pull_request_reviews.required_approving_review_count' || echo "0")
    if [ "$required_approvals" -ge 1 ]; then
        check "Main branch requires at least 1 approval" true
        log_verbose "  Required approvals: $required_approvals"
    else
        check "Main branch requires at least 1 approval" false
    fi

    check "Main branch dismisses stale reviews" \
        gh api "repos/$REPO_FULL/branches/main/protection" | jq -e '.required_pull_request_reviews.dismiss_stale_reviews == true'

    check "Main branch requires status checks" \
        gh api "repos/$REPO_FULL/branches/main/protection" | jq -e '.required_status_checks'

    check "Main branch requires up-to-date branches" \
        gh api "repos/$REPO_FULL/branches/main/protection" | jq -e '.required_status_checks.strict == true'

    # Check for required status checks
    local status_checks
    status_checks=$(gh api "repos/$REPO_FULL/branches/main/protection" 2>/dev/null | jq -r '.required_status_checks.contexts[]' 2>/dev/null || echo "")
    if [ -n "$status_checks" ]; then
        check "Main branch has required status checks configured" true
        log_verbose "  Status checks configured:"
        echo "$status_checks" | while read -r check_name; do
            log_verbose "    - $check_name"
        done
    else
        check "Main branch has required status checks configured" false
    fi

    check "Main branch requires linear history" \
        gh api "repos/$REPO_FULL/branches/main/protection" | jq -e '.required_linear_history.enabled == true'

    check "Main branch prevents force pushes" \
        gh api "repos/$REPO_FULL/branches/main/protection" | jq -e '.allow_force_pushes.enabled == false'

    check "Main branch prevents deletions" \
        gh api "repos/$REPO_FULL/branches/main/protection" | jq -e '.allow_deletions.enabled == false'

    warn "Main branch enforces rules on admins" \
        gh api "repos/$REPO_FULL/branches/main/protection" | jq -e '.enforce_admins.enabled == true'

    warn "Main branch requires signed commits" \
        gh api "repos/$REPO_FULL/branches/main/protection" | jq -e '.required_signatures.enabled == true'

    # Dev branch protection
    log_verbose "Checking dev branch protection..."

    check "Dev branch is protected" \
        gh api "repos/$REPO_FULL/branches/dev/protection"

    check "Dev branch requires pull request reviews" \
        gh api "repos/$REPO_FULL/branches/dev/protection" | jq -e '.required_pull_request_reviews'

    check "Dev branch requires status checks" \
        gh api "repos/$REPO_FULL/branches/dev/protection" | jq -e '.required_status_checks'

    echo
}

#######################################################################
# Repository Settings Validation
#######################################################################

validate_repository_settings() {
    log_info "Validating repository settings..."

    check "Repository allows squash merging" \
        gh api "repos/$REPO_FULL" | jq -e '.allow_squash_merge == true'

    check "Repository allows merge commits" \
        gh api "repos/$REPO_FULL" | jq -e '.allow_merge_commit == true'

    check "Repository disables rebase merging" \
        gh api "repos/$REPO_FULL" | jq -e '.allow_rebase_merge == false'

    check "Repository automatically deletes head branches" \
        gh api "repos/$REPO_FULL" | jq -e '.delete_branch_on_merge == true'

    check "Vulnerability alerts enabled" \
        gh api "repos/$REPO_FULL/vulnerability-alerts"

    warn "Dependabot security updates enabled" \
        gh api "repos/$REPO_FULL/automated-security-fixes"

    warn "Secret scanning enabled" \
        gh api "repos/$REPO_FULL" | jq -e '.security_and_analysis.secret_scanning.status == "enabled"'

    warn "Secret scanning push protection enabled" \
        gh api "repos/$REPO_FULL" | jq -e '.security_and_analysis.secret_scanning_push_protection.status == "enabled"'

    echo
}

#######################################################################
# Workflow Validation
#######################################################################

validate_workflows() {
    log_info "Validating workflows..."

    local workflows=(
        "ci.yml:CI workflow exists"
        "deploy-staging.yml:Staging deployment workflow exists"
        "deploy-production.yml:Production deployment workflow exists"
        "security.yml:Security scanning workflow exists"
        "codeql.yml:CodeQL analysis workflow exists"
    )

    for workflow_check in "${workflows[@]}"; do
        IFS=':' read -r workflow_file description <<< "$workflow_check"
        if [ -f "$PROJECT_ROOT/.github/workflows/$workflow_file" ]; then
            check "$description" true
        else
            warn "$description" false
        fi
    done

    echo
}

#######################################################################
# Generate Report
#######################################################################

generate_json_report() {
    cat <<EOF
{
  "repository": "$REPO_FULL",
  "timestamp": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "summary": {
    "total_checks": $TOTAL_CHECKS,
    "passed": $PASSED_CHECKS,
    "failed": $FAILED_CHECKS,
    "warnings": $WARNING_CHECKS
  },
  "failures": $(printf '%s\n' "${FAILURES[@]}" | jq -R . | jq -s .),
  "warnings": $(printf '%s\n' "${WARNINGS[@]}" | jq -R . | jq -s .),
  "status": "$([ $FAILED_CHECKS -eq 0 ] && echo "pass" || echo "fail")"
}
EOF
}

generate_text_report() {
    echo
    echo "=========================================="
    echo "  Validation Summary"
    echo "=========================================="
    echo "Repository: $REPO_FULL"
    echo "Total checks: $TOTAL_CHECKS"
    echo -e "${GREEN}Passed: $PASSED_CHECKS${NC}"
    [ $FAILED_CHECKS -gt 0 ] && echo -e "${RED}Failed: $FAILED_CHECKS${NC}"
    [ $WARNING_CHECKS -gt 0 ] && echo -e "${YELLOW}Warnings: $WARNING_CHECKS${NC}"
    echo

    if [ $FAILED_CHECKS -gt 0 ]; then
        echo -e "${RED}Failed checks:${NC}"
        for failure in "${FAILURES[@]}"; do
            echo "  ✗ $failure"
        done
        echo
    fi

    if [ $WARNING_CHECKS -gt 0 ]; then
        echo -e "${YELLOW}Warnings:${NC}"
        for warning in "${WARNINGS[@]}"; do
            echo "  ⚠ $warning"
        done
        echo
    fi

    if [ $FAILED_CHECKS -eq 0 ]; then
        echo -e "${GREEN}✓ All required checks passed!${NC}"
        echo
        echo "Next steps:"
        echo "1. Review any warnings above"
        echo "2. Configure missing optional features"
        echo "3. Test deployment workflows"
    else
        echo -e "${RED}✗ Some checks failed${NC}"
        echo
        echo "To fix:"
        echo "1. Review failed checks above"
        echo "2. See docs/GITHUB_CONFIGURATION_GUIDE.md for setup instructions"
        echo "3. Run ./scripts/github-setup.sh to configure automatically"
        echo "4. Run this script again to verify"
    fi
}

#######################################################################
# Main
#######################################################################

show_help() {
    cat << EOF
GitHub Configuration Validation Script

Validates GitHub repository configuration including environments, secrets,
branch protection rules, and repository settings.

Usage:
  $0 [OPTIONS]

Options:
  --environments      Check only environments
  --secrets          Check only secrets
  --branch-protection Check only branch protection
  --all              Check everything (default)
  --verbose          Show detailed output
  --json             Output results as JSON
  --help             Show this help message

Exit Codes:
  0 - All required checks passed
  1 - Some required checks failed
  2 - Critical error (prerequisites not met)

Examples:
  # Run all checks
  $0

  # Check only environments
  $0 --environments

  # Verbose output
  $0 --verbose

  # JSON output for CI/CD
  $0 --json

Requirements:
  - GitHub CLI (gh) installed and authenticated
  - jq for JSON processing

For more information, see:
  docs/GITHUB_CONFIGURATION_GUIDE.md

EOF
}

main() {
    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --environments)
                CHECK_ENVIRONMENTS=true
                CHECK_SECRETS=false
                CHECK_BRANCH_PROTECTION=false
                shift
                ;;
            --secrets)
                CHECK_ENVIRONMENTS=false
                CHECK_SECRETS=true
                CHECK_BRANCH_PROTECTION=false
                shift
                ;;
            --branch-protection)
                CHECK_ENVIRONMENTS=false
                CHECK_SECRETS=false
                CHECK_BRANCH_PROTECTION=true
                shift
                ;;
            --all)
                CHECK_ENVIRONMENTS=true
                CHECK_SECRETS=true
                CHECK_BRANCH_PROTECTION=true
                shift
                ;;
            --verbose)
                VERBOSE=true
                shift
                ;;
            --json)
                JSON_OUTPUT=true
                shift
                ;;
            --help)
                show_help
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                show_help
                exit 2
                ;;
        esac
    done

    # Print banner (unless JSON output)
    if [ "$JSON_OUTPUT" = false ]; then
        echo
        log_info "=========================================="
        log_info "  GitHub Configuration Validation"
        log_info "=========================================="
        log_info "Repository: $REPO_FULL"
        echo
    fi

    # Check prerequisites
    check_prerequisites

    # Run validations
    validate_environments
    validate_secrets
    validate_branch_protection
    validate_repository_settings
    validate_workflows

    # Generate report
    if [ "$JSON_OUTPUT" = true ]; then
        generate_json_report
    else
        generate_text_report
    fi

    # Exit with appropriate code
    if [ $FAILED_CHECKS -gt 0 ]; then
        exit 1
    else
        exit 0
    fi
}

# Run main function
main "$@"
