# Repository-level secrets
# Note: Only manage secrets via Terraform if you're comfortable storing them in state
# Consider using GitHub UI or gh CLI for sensitive secrets

# Codecov token (optional)
resource "github_actions_secret" "codecov_token" {
  count = var.manage_secrets && var.codecov_token != "" ? 1 : 0

  repository      = github_repository.main.name
  secret_name     = "CODECOV_TOKEN"
  plaintext_value = var.codecov_token
}

# Infrastructure-specific secrets for Kubernetes
resource "github_actions_secret" "kubeconfig" {
  count = var.manage_secrets && var.deployment_type == "kubernetes" && var.kubeconfig != "" ? 1 : 0

  repository      = github_repository.main.name
  secret_name     = "KUBECONFIG"
  plaintext_value = var.kubeconfig
}

resource "github_actions_secret" "k8s_cluster_url" {
  count = var.manage_secrets && var.deployment_type == "kubernetes" && var.k8s_cluster_url != "" ? 1 : 0

  repository      = github_repository.main.name
  secret_name     = "K8S_CLUSTER_URL"
  plaintext_value = var.k8s_cluster_url
}

resource "github_actions_secret" "k8s_token" {
  count = var.manage_secrets && var.deployment_type == "kubernetes" && var.k8s_token != "" ? 1 : 0

  repository      = github_repository.main.name
  secret_name     = "K8S_TOKEN"
  plaintext_value = var.k8s_token
}

resource "github_actions_secret" "k8s_namespace" {
  count = var.manage_secrets && var.deployment_type == "kubernetes" && var.k8s_namespace != "" ? 1 : 0

  repository      = github_repository.main.name
  secret_name     = "K8S_NAMESPACE"
  plaintext_value = var.k8s_namespace
}

# Infrastructure-specific secrets for AWS ECS
resource "github_actions_secret" "aws_access_key_id" {
  count = var.manage_secrets && var.deployment_type == "aws-ecs" && var.aws_access_key_id != "" ? 1 : 0

  repository      = github_repository.main.name
  secret_name     = "AWS_ACCESS_KEY_ID"
  plaintext_value = var.aws_access_key_id
}

resource "github_actions_secret" "aws_secret_access_key" {
  count = var.manage_secrets && var.deployment_type == "aws-ecs" && var.aws_secret_access_key != "" ? 1 : 0

  repository      = github_repository.main.name
  secret_name     = "AWS_SECRET_ACCESS_KEY"
  plaintext_value = var.aws_secret_access_key
}

resource "github_actions_secret" "aws_region" {
  count = var.manage_secrets && var.deployment_type == "aws-ecs" && var.aws_region != "" ? 1 : 0

  repository      = github_repository.main.name
  secret_name     = "AWS_REGION"
  plaintext_value = var.aws_region
}

resource "github_actions_secret" "ecs_cluster" {
  count = var.manage_secrets && var.deployment_type == "aws-ecs" && var.ecs_cluster != "" ? 1 : 0

  repository      = github_repository.main.name
  secret_name     = "ECS_CLUSTER"
  plaintext_value = var.ecs_cluster
}

resource "github_actions_secret" "ecs_service" {
  count = var.manage_secrets && var.deployment_type == "aws-ecs" && var.ecs_service != "" ? 1 : 0

  repository      = github_repository.main.name
  secret_name     = "ECS_SERVICE"
  plaintext_value = var.ecs_service
}

resource "github_actions_secret" "ecs_task_definition" {
  count = var.manage_secrets && var.deployment_type == "aws-ecs" && var.ecs_task_definition != "" ? 1 : 0

  repository      = github_repository.main.name
  secret_name     = "ECS_TASK_DEFINITION"
  plaintext_value = var.ecs_task_definition
}

# Infrastructure-specific secrets for SSH
resource "github_actions_secret" "ssh_private_key" {
  count = var.manage_secrets && var.deployment_type == "ssh" && var.ssh_private_key != "" ? 1 : 0

  repository      = github_repository.main.name
  secret_name     = "SSH_PRIVATE_KEY"
  plaintext_value = var.ssh_private_key
}

resource "github_actions_secret" "ssh_host" {
  count = var.manage_secrets && var.deployment_type == "ssh" && var.ssh_host != "" ? 1 : 0

  repository      = github_repository.main.name
  secret_name     = "SSH_HOST"
  plaintext_value = var.ssh_host
}

resource "github_actions_secret" "ssh_user" {
  count = var.manage_secrets && var.deployment_type == "ssh" && var.ssh_user != "" ? 1 : 0

  repository      = github_repository.main.name
  secret_name     = "SSH_USER"
  plaintext_value = var.ssh_user
}

resource "github_actions_secret" "ssh_port" {
  count = var.manage_secrets && var.deployment_type == "ssh" && var.ssh_port != "" ? 1 : 0

  repository      = github_repository.main.name
  secret_name     = "SSH_PORT"
  plaintext_value = var.ssh_port
}

# Repository variables (non-sensitive)
resource "github_actions_variable" "registry" {
  repository    = github_repository.main.name
  variable_name = "REGISTRY"
  value         = var.container_registry
}

resource "github_actions_variable" "image_name" {
  repository    = github_repository.main.name
  variable_name = "IMAGE_NAME"
  value         = var.container_image_name
}

# Outputs
output "configured_secrets" {
  description = "List of configured secrets (names only, not values)"
  value = concat(
    var.manage_secrets && var.codecov_token != "" ? ["CODECOV_TOKEN"] : [],
    var.manage_secrets && var.deployment_type == "kubernetes" ? [
      var.kubeconfig != "" ? "KUBECONFIG" : "",
      var.k8s_cluster_url != "" ? "K8S_CLUSTER_URL" : "",
      var.k8s_token != "" ? "K8S_TOKEN" : "",
      var.k8s_namespace != "" ? "K8S_NAMESPACE" : "",
    ] : [],
    var.manage_secrets && var.deployment_type == "aws-ecs" ? [
      var.aws_access_key_id != "" ? "AWS_ACCESS_KEY_ID" : "",
      var.aws_secret_access_key != "" ? "AWS_SECRET_ACCESS_KEY" : "",
      var.aws_region != "" ? "AWS_REGION" : "",
      var.ecs_cluster != "" ? "ECS_CLUSTER" : "",
      var.ecs_service != "" ? "ECS_SERVICE" : "",
      var.ecs_task_definition != "" ? "ECS_TASK_DEFINITION" : "",
    ] : [],
    var.manage_secrets && var.deployment_type == "ssh" ? [
      var.ssh_private_key != "" ? "SSH_PRIVATE_KEY" : "",
      var.ssh_host != "" ? "SSH_HOST" : "",
      var.ssh_user != "" ? "SSH_USER" : "",
      var.ssh_port != "" ? "SSH_PORT" : "",
    ] : [],
  )
}
