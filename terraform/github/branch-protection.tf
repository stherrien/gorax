# Main branch protection
resource "github_branch_protection" "main" {
  repository_id = github_repository.main.node_id
  pattern       = "main"

  # Require pull request reviews
  required_pull_request_reviews {
    dismiss_stale_reviews           = true
    require_code_owner_reviews      = true
    required_approving_review_count = 1
    require_last_push_approval      = true
    restrict_dismissals             = false
  }

  # Require status checks
  required_status_checks {
    strict = true
    contexts = [
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
      "Secrets Scanning / gitleaks",
    ]
  }

  # Enforce admins
  enforce_admins = true

  # Require signed commits
  require_signed_commits = var.require_signed_commits

  # Require linear history
  require_linear_history = true

  # Restrict pushes
  restrict_pushes {
    blocks_creations = true
    push_allowances  = var.main_branch_push_allowances
  }

  # Prevent force pushes and deletions
  allows_force_pushes = false
  allows_deletions    = false

  # Require conversation resolution
  require_conversation_resolution = var.require_conversation_resolution

  # Lock branch (set to false for normal operations)
  lock_branch = false
}

# Dev branch protection
resource "github_branch_protection" "dev" {
  repository_id = github_repository.main.node_id
  pattern       = "dev"

  # Require pull request reviews
  required_pull_request_reviews {
    dismiss_stale_reviews           = true
    require_code_owner_reviews      = false
    required_approving_review_count = 1
    require_last_push_approval      = false
    restrict_dismissals             = false
  }

  # Require status checks
  required_status_checks {
    strict = true
    contexts = [
      "Go Tests",
      "Go Lint",
      "Frontend Tests",
      "Frontend Lint",
      "Coverage Threshold Check",
      "Build Verification",
      "Security Scanning / gosec",
      "Security Scanning / npm-audit",
    ]
  }

  # Don't enforce on admins for dev branch
  enforce_admins = false

  # Signed commits optional for dev
  require_signed_commits = false

  # Linear history optional for dev
  require_linear_history = false

  # Restrict pushes
  restrict_pushes {
    blocks_creations = true
    push_allowances  = var.dev_branch_push_allowances
  }

  # Prevent force pushes and deletions
  allows_force_pushes = false
  allows_deletions    = false

  # Require conversation resolution
  require_conversation_resolution = false

  # Lock branch
  lock_branch = false
}

# Release branches protection (pattern-based)
resource "github_branch_protection" "release" {
  count = var.enable_release_branch_protection ? 1 : 0

  repository_id = github_repository.main.node_id
  pattern       = "release/*"

  # Require pull request reviews
  required_pull_request_reviews {
    dismiss_stale_reviews           = true
    require_code_owner_reviews      = true
    required_approving_review_count = 2
    require_last_push_approval      = true
    restrict_dismissals             = false
  }

  # Require status checks
  required_status_checks {
    strict = true
    contexts = [
      "Go Tests",
      "Go Lint",
      "Frontend Tests",
      "Frontend Lint",
      "Coverage Threshold Check",
      "Build Verification",
    ]
  }

  # Enforce admins
  enforce_admins = true

  # Require signed commits
  require_signed_commits = var.require_signed_commits

  # Require linear history
  require_linear_history = true

  # Restrict pushes
  restrict_pushes {
    blocks_creations = true
    push_allowances  = var.release_branch_push_allowances
  }

  # Prevent force pushes and deletions
  allows_force_pushes = false
  allows_deletions    = false

  # Lock branch
  lock_branch = false
}

# Hotfix branches protection (pattern-based)
resource "github_branch_protection" "hotfix" {
  count = var.enable_hotfix_branch_protection ? 1 : 0

  repository_id = github_repository.main.node_id
  pattern       = "hotfix/*"

  # Require pull request reviews (relaxed for emergencies)
  required_pull_request_reviews {
    dismiss_stale_reviews           = true
    require_code_owner_reviews      = false
    required_approving_review_count = 1
    require_last_push_approval      = false
    restrict_dismissals             = false
  }

  # Require critical status checks only
  required_status_checks {
    strict = true
    contexts = [
      "Go Tests",
      "Frontend Tests",
      "Security Scanning / gosec",
    ]
  }

  # Don't enforce on admins for hotfixes
  enforce_admins = false

  # Signed commits optional for hotfixes
  require_signed_commits = false

  # Linear history optional for hotfixes
  require_linear_history = false

  # Allow force pushes for hotfixes (emergency only)
  allows_force_pushes = false
  allows_deletions    = false

  # Lock branch
  lock_branch = false
}

# Outputs
output "main_branch_protection_url" {
  description = "URL to main branch protection settings"
  value       = "https://github.com/${var.github_owner}/${var.repository_name}/settings/branches"
}

output "protected_branches" {
  description = "List of protected branch patterns"
  value = concat(
    ["main", "dev"],
    var.enable_release_branch_protection ? ["release/*"] : [],
    var.enable_hotfix_branch_protection ? ["hotfix/*"] : []
  )
}
