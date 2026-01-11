terraform {
  required_version = ">= 1.0.0"

  required_providers {
    github = {
      source  = "integrations/github"
      version = "~> 6.0"
    }
  }

  # Uncomment to use remote state
  # backend "s3" {
  #   bucket         = "gorax-terraform-state"
  #   key            = "github/terraform.tfstate"
  #   region         = "us-east-1"
  #   encrypt        = true
  #   dynamodb_table = "gorax-terraform-locks"
  # }
}

provider "github" {
  owner = var.github_owner
  token = var.github_token
}

# Repository configuration
resource "github_repository" "main" {
  name        = var.repository_name
  description = "Enterprise workflow automation platform with AI-powered workflow generation"

  visibility = "public"

  # Repository features
  has_issues      = true
  has_projects    = true
  has_wiki        = false
  has_downloads   = true
  has_discussions = true

  # Security features
  vulnerability_alerts   = true
  delete_branch_on_merge = true

  # Topics
  topics = [
    "workflow-automation",
    "golang",
    "react",
    "typescript",
    "low-code",
    "ai-powered",
    "enterprise"
  ]

  # Homepage
  homepage_url = var.homepage_url

  # Security and analysis
  security_and_analysis {
    secret_scanning {
      status = "enabled"
    }
    secret_scanning_push_protection {
      status = "enabled"
    }
  }

  # Merge settings
  allow_squash_merge     = true
  allow_merge_commit     = true
  allow_rebase_merge     = false
  allow_auto_merge       = false
  squash_merge_commit_title   = "PR_TITLE"
  squash_merge_commit_message = "PR_BODY"
}

# Repository settings
resource "github_repository_dependabot_security_updates" "main" {
  repository = github_repository.main.name
  enabled    = true
}

# Default labels
resource "github_issue_label" "priority_critical" {
  repository  = github_repository.main.name
  name        = "priority: critical"
  color       = "d73a4a"
  description = "Critical priority issue"
}

resource "github_issue_label" "priority_high" {
  repository  = github_repository.main.name
  name        = "priority: high"
  color       = "ff6b6b"
  description = "High priority issue"
}

resource "github_issue_label" "priority_medium" {
  repository  = github_repository.main.name
  name        = "priority: medium"
  color       = "ffa500"
  description = "Medium priority issue"
}

resource "github_issue_label" "priority_low" {
  repository  = github_repository.main.name
  name        = "priority: low"
  color       = "90ee90"
  description = "Low priority issue"
}

resource "github_issue_label" "type_bug" {
  repository  = github_repository.main.name
  name        = "type: bug"
  color       = "d73a4a"
  description = "Bug or defect"
}

resource "github_issue_label" "type_feature" {
  repository  = github_repository.main.name
  name        = "type: feature"
  color       = "a2eeef"
  description = "New feature request"
}

resource "github_issue_label" "type_enhancement" {
  repository  = github_repository.main.name
  name        = "type: enhancement"
  color       = "84b6eb"
  description = "Enhancement to existing feature"
}

resource "github_issue_label" "type_documentation" {
  repository  = github_repository.main.name
  name        = "type: documentation"
  color       = "0075ca"
  description = "Documentation improvements"
}

resource "github_issue_label" "type_security" {
  repository  = github_repository.main.name
  name        = "type: security"
  color       = "ee0701"
  description = "Security-related issue"
}

# Repository collaborators (example)
# Uncomment and customize as needed
# resource "github_repository_collaborator" "collaborator" {
#   repository = github_repository.main.name
#   username   = "collaborator-username"
#   permission = "push"
# }

# Team repository access (if using organization)
# resource "github_team_repository" "team_access" {
#   team_id    = var.team_id
#   repository = github_repository.main.name
#   permission = "push"
# }

# Webhooks (example for Slack)
# resource "github_repository_webhook" "slack" {
#   repository = github_repository.main.name
#
#   configuration {
#     url          = var.slack_webhook_url
#     content_type = "json"
#     insecure_ssl = false
#   }
#
#   active = true
#
#   events = [
#     "push",
#     "pull_request",
#     "issues",
#     "release",
#   ]
# }

# Outputs
output "repository_full_name" {
  description = "Full name of the repository"
  value       = github_repository.main.full_name
}

output "repository_html_url" {
  description = "URL of the repository"
  value       = github_repository.main.html_url
}

output "repository_ssh_url" {
  description = "SSH URL of the repository"
  value       = github_repository.main.ssh_clone_url
}

output "repository_http_url" {
  description = "HTTPS URL of the repository"
  value       = github_repository.main.http_clone_url
}
