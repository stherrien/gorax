variable "github_owner" {
  description = "GitHub organization or user name"
  type        = string
  default     = "stherrien"
}

variable "github_token" {
  description = "GitHub personal access token with repo and admin:org permissions"
  type        = string
  sensitive   = true
}

variable "repository_name" {
  description = "Name of the GitHub repository"
  type        = string
  default     = "gorax"
}

variable "homepage_url" {
  description = "Repository homepage URL"
  type        = string
  default     = "https://gorax.dev"
}

# Environment configuration
variable "staging_url" {
  description = "Staging environment URL"
  type        = string
  default     = "https://staging.gorax.dev"
}

variable "production_url" {
  description = "Production environment URL"
  type        = string
  default     = "https://gorax.dev"
}

variable "production_reviewers" {
  description = "List of GitHub user IDs who can approve production deployments"
  type        = list(number)
  default     = []
}

variable "production_reviewer_teams" {
  description = "List of GitHub team IDs who can approve production deployments"
  type        = list(number)
  default     = []
}

variable "enable_development_environment" {
  description = "Whether to create a development environment"
  type        = bool
  default     = false
}

# Secret management
variable "manage_secrets" {
  description = "Whether to manage secrets via Terraform (WARNING: secrets will be stored in state)"
  type        = bool
  default     = false
}

variable "codecov_token" {
  description = "Codecov upload token"
  type        = string
  default     = ""
  sensitive   = true
}

variable "staging_db_password" {
  description = "Staging database password"
  type        = string
  default     = ""
  sensitive   = true
}

variable "staging_redis_password" {
  description = "Staging Redis password"
  type        = string
  default     = ""
  sensitive   = true
}

variable "production_db_password" {
  description = "Production database password"
  type        = string
  default     = ""
  sensitive   = true
}

variable "production_redis_password" {
  description = "Production Redis password"
  type        = string
  default     = ""
  sensitive   = true
}

variable "health_check_token" {
  description = "Token for health check endpoints"
  type        = string
  default     = ""
  sensitive   = true
}

# Deployment configuration
variable "deployment_type" {
  description = "Type of deployment infrastructure (kubernetes, aws-ecs, ssh)"
  type        = string
  default     = "kubernetes"
  validation {
    condition     = contains(["kubernetes", "aws-ecs", "ssh"], var.deployment_type)
    error_message = "deployment_type must be one of: kubernetes, aws-ecs, ssh"
  }
}

# Kubernetes secrets
variable "kubeconfig" {
  description = "Base64-encoded kubeconfig file"
  type        = string
  default     = ""
  sensitive   = true
}

variable "k8s_cluster_url" {
  description = "Kubernetes cluster API URL"
  type        = string
  default     = ""
}

variable "k8s_token" {
  description = "Kubernetes service account token"
  type        = string
  default     = ""
  sensitive   = true
}

variable "k8s_namespace" {
  description = "Kubernetes namespace for deployments"
  type        = string
  default     = "gorax"
}

# AWS ECS secrets
variable "aws_access_key_id" {
  description = "AWS access key ID"
  type        = string
  default     = ""
  sensitive   = true
}

variable "aws_secret_access_key" {
  description = "AWS secret access key"
  type        = string
  default     = ""
  sensitive   = true
}

variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

variable "ecs_cluster" {
  description = "ECS cluster name"
  type        = string
  default     = ""
}

variable "ecs_service" {
  description = "ECS service name"
  type        = string
  default     = ""
}

variable "ecs_task_definition" {
  description = "ECS task definition family"
  type        = string
  default     = ""
}

# SSH secrets
variable "ssh_private_key" {
  description = "SSH private key for deployment"
  type        = string
  default     = ""
  sensitive   = true
}

variable "ssh_host" {
  description = "SSH host for deployment"
  type        = string
  default     = ""
}

variable "ssh_user" {
  description = "SSH user for deployment"
  type        = string
  default     = "deploy"
}

variable "ssh_port" {
  description = "SSH port for deployment"
  type        = string
  default     = "22"
}

# Branch protection
variable "require_signed_commits" {
  description = "Whether to require signed commits on protected branches"
  type        = bool
  default     = true
}

variable "require_conversation_resolution" {
  description = "Whether to require conversation resolution before merging"
  type        = bool
  default     = false
}

variable "main_branch_push_allowances" {
  description = "List of actors allowed to push to main branch (format: /user_id or /team_id)"
  type        = list(string)
  default     = []
}

variable "dev_branch_push_allowances" {
  description = "List of actors allowed to push to dev branch (format: /user_id or /team_id)"
  type        = list(string)
  default     = []
}

variable "release_branch_push_allowances" {
  description = "List of actors allowed to push to release branches (format: /user_id or /team_id)"
  type        = list(string)
  default     = []
}

variable "enable_release_branch_protection" {
  description = "Whether to enable protection for release/* branches"
  type        = bool
  default     = false
}

variable "enable_hotfix_branch_protection" {
  description = "Whether to enable protection for hotfix/* branches"
  type        = bool
  default     = false
}

# Container registry
variable "container_registry" {
  description = "Container registry URL"
  type        = string
  default     = "ghcr.io"
}

variable "container_image_name" {
  description = "Container image name"
  type        = string
  default     = "stherrien/gorax"
}

# Webhook configuration
variable "slack_webhook_url" {
  description = "Slack webhook URL for notifications"
  type        = string
  default     = ""
  sensitive   = true
}
