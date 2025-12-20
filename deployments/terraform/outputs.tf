# Root Terraform Outputs

# VPC Outputs
output "vpc_id" {
  description = "ID of the VPC"
  value       = module.vpc.vpc_id
}

output "private_subnet_ids" {
  description = "IDs of private subnets"
  value       = module.vpc.private_subnet_ids
}

output "public_subnet_ids" {
  description = "IDs of public subnets"
  value       = module.vpc.public_subnet_ids
}

# EKS Outputs
output "eks_cluster_name" {
  description = "Name of the EKS cluster"
  value       = module.eks.cluster_name
}

output "eks_cluster_endpoint" {
  description = "Endpoint for EKS cluster API"
  value       = module.eks.cluster_endpoint
}

output "eks_cluster_certificate_authority_data" {
  description = "Base64 encoded certificate data for cluster"
  value       = module.eks.cluster_certificate_authority_data
  sensitive   = true
}

output "eks_oidc_provider_arn" {
  description = "ARN of the OIDC provider for IRSA"
  value       = module.eks.oidc_provider_arn
}

# Database Outputs
output "database_endpoint" {
  description = "Aurora cluster endpoint"
  value       = module.aurora.cluster_endpoint
}

output "database_reader_endpoint" {
  description = "Aurora cluster reader endpoint"
  value       = module.aurora.cluster_reader_endpoint
}

output "database_port" {
  description = "Aurora cluster port"
  value       = module.aurora.cluster_port
}

output "database_secret_arn" {
  description = "ARN of Secrets Manager secret for database credentials"
  value       = module.aurora.secret_arn
}

# Redis Outputs
output "redis_primary_endpoint" {
  description = "Redis primary endpoint address"
  value       = module.elasticache.primary_endpoint_address
}

output "redis_reader_endpoint" {
  description = "Redis reader endpoint address"
  value       = module.elasticache.reader_endpoint_address
}

output "redis_port" {
  description = "Redis port"
  value       = module.elasticache.port
}

output "redis_secret_arn" {
  description = "ARN of Secrets Manager secret for Redis credentials"
  value       = module.elasticache.secret_arn
}

# SQS Outputs
output "main_queue_url" {
  description = "URL of the main workflow queue"
  value       = module.sqs.main_queue_id
}

output "main_queue_arn" {
  description = "ARN of the main workflow queue"
  value       = module.sqs.main_queue_arn
}

output "high_priority_queue_url" {
  description = "URL of the high priority queue"
  value       = module.sqs.high_priority_queue_id
}

output "high_priority_queue_arn" {
  description = "ARN of the high priority queue"
  value       = module.sqs.high_priority_queue_arn
}

# S3 Outputs
output "artifacts_bucket_name" {
  description = "Name of the artifacts S3 bucket"
  value       = module.s3.artifacts_bucket_id
}

output "artifacts_bucket_arn" {
  description = "ARN of the artifacts S3 bucket"
  value       = module.s3.artifacts_bucket_arn
}

output "logs_bucket_name" {
  description = "Name of the logs S3 bucket"
  value       = module.s3.logs_bucket_id
}

output "logs_bucket_arn" {
  description = "ARN of the logs S3 bucket"
  value       = module.s3.logs_bucket_arn
}

# IAM Outputs
output "api_service_role_arn" {
  description = "ARN of the API service IAM role"
  value       = module.iam.api_service_role_arn
}

output "worker_service_role_arn" {
  description = "ARN of the worker service IAM role"
  value       = module.iam.worker_service_role_arn
}

# KMS Outputs
output "database_kms_key_arn" {
  description = "ARN of the database encryption KMS key"
  value       = module.kms.database_key_arn
}

output "secrets_kms_key_arn" {
  description = "ARN of the Secrets Manager KMS key"
  value       = module.kms.secrets_key_arn
}

# Kubernetes Configuration
output "kubectl_config_command" {
  description = "Command to configure kubectl"
  value       = "aws eks update-kubeconfig --region ${var.aws_region} --name ${module.eks.cluster_name}"
}

# Service Account Annotations
output "api_service_account_annotation" {
  description = "Annotation for API service account"
  value       = "eks.amazonaws.com/role-arn: ${module.iam.api_service_role_arn}"
}

output "worker_service_account_annotation" {
  description = "Annotation for worker service account"
  value       = "eks.amazonaws.com/role-arn: ${module.iam.worker_service_role_arn}"
}
