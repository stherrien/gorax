# Staging environment
resource "github_repository_environment" "staging" {
  repository  = github_repository.main.name
  environment = "staging"

  # No deployment protection for staging (auto-deploy from dev)
  deployment_branch_policy {
    protected_branches     = false
    custom_branch_policies = true
  }
}

# Staging deployment branch policy
resource "github_repository_environment_deployment_policy" "staging_dev" {
  repository     = github_repository.main.name
  environment    = github_repository_environment.staging.environment
  branch_pattern = "dev"
}

# Production environment
resource "github_repository_environment" "production" {
  repository  = github_repository.main.name
  environment = "production"

  # Reviewers required for production deployments
  reviewers {
    users = var.production_reviewers
    # teams = var.production_reviewer_teams  # Uncomment if using teams
  }

  # Wait timer (5 minutes)
  wait_timer = 5

  # Deployment branch policy
  deployment_branch_policy {
    protected_branches     = false
    custom_branch_policies = true
  }
}

# Production deployment branch policy
resource "github_repository_environment_deployment_policy" "production_main" {
  repository     = github_repository.main.name
  environment    = github_repository_environment.production.environment
  branch_pattern = "main"
}

# Development environment (optional, for local testing)
resource "github_repository_environment" "development" {
  count = var.enable_development_environment ? 1 : 0

  repository  = github_repository.main.name
  environment = "development"

  # No protection rules for development
  deployment_branch_policy {
    protected_branches     = false
    custom_branch_policies = false
  }
}

# Environment secrets (placeholders)
# Note: Actual secret values should be set via GitHub UI or gh CLI
# Terraform can manage secret existence but not view values

# Staging environment secrets
resource "github_actions_environment_secret" "staging_url" {
  repository      = github_repository.main.name
  environment     = github_repository_environment.staging.environment
  secret_name     = "STAGING_URL"
  plaintext_value = var.staging_url
}

resource "github_actions_environment_secret" "staging_db_password" {
  count = var.manage_secrets ? 1 : 0

  repository      = github_repository.main.name
  environment     = github_repository_environment.staging.environment
  secret_name     = "STAGING_DB_PASSWORD"
  plaintext_value = var.staging_db_password
}

resource "github_actions_environment_secret" "staging_redis_password" {
  count = var.manage_secrets ? 1 : 0

  repository      = github_repository.main.name
  environment     = github_repository_environment.staging.environment
  secret_name     = "STAGING_REDIS_PASSWORD"
  plaintext_value = var.staging_redis_password
}

# Production environment secrets
resource "github_actions_environment_secret" "production_url" {
  repository      = github_repository.main.name
  environment     = github_repository_environment.production.environment
  secret_name     = "PRODUCTION_URL"
  plaintext_value = var.production_url
}

resource "github_actions_environment_secret" "production_db_password" {
  count = var.manage_secrets ? 1 : 0

  repository      = github_repository.main.name
  environment     = github_repository_environment.production.environment
  secret_name     = "PRODUCTION_DB_PASSWORD"
  plaintext_value = var.production_db_password
}

resource "github_actions_environment_secret" "production_redis_password" {
  count = var.manage_secrets ? 1 : 0

  repository      = github_repository.main.name
  environment     = github_repository_environment.production.environment
  secret_name     = "PRODUCTION_REDIS_PASSWORD"
  plaintext_value = var.production_redis_password
}

resource "github_actions_environment_secret" "health_check_token" {
  count = var.manage_secrets ? 1 : 0

  repository      = github_repository.main.name
  environment     = github_repository_environment.production.environment
  secret_name     = "HEALTH_CHECK_TOKEN"
  plaintext_value = var.health_check_token
}

# Environment variables (non-sensitive)
resource "github_actions_environment_variable" "staging_environment" {
  repository    = github_repository.main.name
  environment   = github_repository_environment.staging.environment
  variable_name = "ENVIRONMENT"
  value         = "staging"
}

resource "github_actions_environment_variable" "staging_log_level" {
  repository    = github_repository.main.name
  environment   = github_repository_environment.staging.environment
  variable_name = "LOG_LEVEL"
  value         = "debug"
}

resource "github_actions_environment_variable" "production_environment" {
  repository    = github_repository.main.name
  environment   = github_repository_environment.production.environment
  variable_name = "ENVIRONMENT"
  value         = "production"
}

resource "github_actions_environment_variable" "production_log_level" {
  repository    = github_repository.main.name
  environment   = github_repository_environment.production.environment
  variable_name = "LOG_LEVEL"
  value         = "info"
}

# Outputs
output "staging_environment_url" {
  description = "Staging environment URL"
  value       = "https://github.com/${var.github_owner}/${var.repository_name}/deployments/activity_log?environment=staging"
}

output "production_environment_url" {
  description = "Production environment URL"
  value       = "https://github.com/${var.github_owner}/${var.repository_name}/deployments/activity_log?environment=production"
}
