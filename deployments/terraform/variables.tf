# Root Terraform Variables

# General Configuration
variable "project_name" {
  description = "Name of the project"
  type        = string
  default     = "gorax"
}

variable "environment" {
  description = "Environment name (dev, staging, production)"
  type        = string
  validation {
    condition     = contains(["dev", "staging", "production"], var.environment)
    error_message = "Environment must be dev, staging, or production."
  }
}

variable "aws_region" {
  description = "AWS region for resources"
  type        = string
  default     = "us-east-1"
}

variable "availability_zone_count" {
  description = "Number of availability zones to use"
  type        = number
  default     = 3
}

variable "tags" {
  description = "Additional tags to apply to all resources"
  type        = map(string)
  default     = {}
}

# VPC Configuration
variable "vpc_cidr" {
  description = "CIDR block for VPC"
  type        = string
  default     = "10.0.0.0/16"
}

variable "enable_vpc_flow_logs" {
  description = "Enable VPC flow logs"
  type        = bool
  default     = true
}

variable "vpc_flow_logs_retention_days" {
  description = "Number of days to retain VPC flow logs"
  type        = number
  default     = 30
}

# KMS Configuration
variable "kms_deletion_window_days" {
  description = "Number of days before KMS keys are deleted"
  type        = number
  default     = 30
}

# EKS Configuration
variable "kubernetes_version" {
  description = "Kubernetes version"
  type        = string
  default     = "1.28"
}

variable "kubernetes_namespace" {
  description = "Kubernetes namespace for Gorax services"
  type        = string
  default     = "gorax"
}

variable "eks_enable_public_access" {
  description = "Enable public access to EKS API"
  type        = bool
  default     = true
}

variable "eks_public_access_cidrs" {
  description = "CIDR blocks allowed to access EKS API publicly"
  type        = list(string)
  default     = ["0.0.0.0/0"]
}

variable "eks_workstation_cidrs" {
  description = "CIDR blocks for workstation access"
  type        = list(string)
  default     = []
}

variable "eks_cluster_log_types" {
  description = "EKS control plane log types to enable"
  type        = list(string)
  default     = ["api", "audit", "authenticator", "controllerManager", "scheduler"]
}

variable "eks_log_retention_days" {
  description = "Number of days to retain EKS logs"
  type        = number
  default     = 30
}

variable "eks_node_groups" {
  description = "EKS node group configurations"
  type = map(object({
    instance_types   = list(string)
    capacity_type    = string
    disk_size        = number
    desired_size     = number
    max_size         = number
    min_size         = number
    max_unavailable  = number
    labels           = map(string)
    taints           = list(object({
      key    = string
      value  = string
      effect = string
    }))
    tags             = map(string)
  }))
  default = {
    general = {
      instance_types  = ["t3.large"]
      capacity_type   = "ON_DEMAND"
      disk_size       = 50
      desired_size    = 2
      max_size        = 5
      min_size        = 1
      max_unavailable = 1
      labels          = {}
      taints          = []
      tags            = {}
    }
  }
}

# Aurora Configuration
variable "database_name" {
  description = "Name of the default database"
  type        = string
  default     = "gorax"
}

variable "database_master_username" {
  description = "Master username for Aurora"
  type        = string
  default     = "gorax_admin"
}

variable "aurora_engine_version" {
  description = "Aurora PostgreSQL engine version"
  type        = string
  default     = "15.4"
}

variable "aurora_engine_family" {
  description = "Aurora PostgreSQL engine family"
  type        = string
  default     = "aurora-postgresql15"
}

variable "aurora_enable_serverless_v2" {
  description = "Enable Aurora Serverless v2"
  type        = bool
  default     = true
}

variable "aurora_serverless_min_capacity" {
  description = "Minimum Aurora Serverless v2 capacity"
  type        = number
  default     = 0.5
}

variable "aurora_serverless_max_capacity" {
  description = "Maximum Aurora Serverless v2 capacity"
  type        = number
  default     = 2
}

variable "aurora_instance_count" {
  description = "Number of Aurora instances"
  type        = number
  default     = 2
}

variable "aurora_backup_retention_period" {
  description = "Number of days to retain Aurora backups"
  type        = number
  default     = 7
}

variable "aurora_deletion_protection" {
  description = "Enable deletion protection for Aurora"
  type        = bool
  default     = true
}

variable "aurora_monitoring_interval" {
  description = "Enhanced monitoring interval for Aurora"
  type        = number
  default     = 60
}

variable "aurora_enable_performance_insights" {
  description = "Enable Performance Insights for Aurora"
  type        = bool
  default     = true
}

# ElastiCache Configuration
variable "redis_engine_version" {
  description = "Redis engine version"
  type        = string
  default     = "7.0"
}

variable "redis_parameter_group_family" {
  description = "Redis parameter group family"
  type        = string
  default     = "redis7"
}

variable "redis_node_type" {
  description = "Redis node instance type"
  type        = string
  default     = "cache.t4g.small"
}

variable "redis_num_cache_nodes" {
  description = "Number of Redis cache nodes"
  type        = number
  default     = 2
}

variable "redis_automatic_failover_enabled" {
  description = "Enable automatic failover for Redis"
  type        = bool
  default     = true
}

variable "redis_multi_az_enabled" {
  description = "Enable Multi-AZ for Redis"
  type        = bool
  default     = true
}

variable "redis_snapshot_retention_limit" {
  description = "Number of days to retain Redis snapshots"
  type        = number
  default     = 7
}

variable "redis_transit_encryption_enabled" {
  description = "Enable encryption in transit for Redis"
  type        = bool
  default     = true
}

# SQS Configuration
variable "sqs_visibility_timeout_seconds" {
  description = "Visibility timeout for main SQS queue"
  type        = number
  default     = 300
}

variable "sqs_high_priority_visibility_timeout_seconds" {
  description = "Visibility timeout for high priority SQS queue"
  type        = number
  default     = 60
}

variable "sqs_message_retention_seconds" {
  description = "Message retention period for SQS queues"
  type        = number
  default     = 345600
}

variable "sqs_max_receive_count" {
  description = "Maximum receives before sending to DLQ"
  type        = number
  default     = 3
}

variable "sqs_create_fifo_queue" {
  description = "Create FIFO queue for ordered processing"
  type        = bool
  default     = false
}

# S3 Configuration
variable "s3_artifacts_transition_to_ia_days" {
  description = "Days before transitioning artifacts to IA"
  type        = number
  default     = 90
}

variable "s3_artifacts_transition_to_glacier_days" {
  description = "Days before transitioning artifacts to Glacier"
  type        = number
  default     = 180
}

variable "s3_logs_transition_to_ia_days" {
  description = "Days before transitioning logs to IA"
  type        = number
  default     = 30
}

variable "s3_logs_transition_to_glacier_days" {
  description = "Days before transitioning logs to Glacier"
  type        = number
  default     = 90
}

variable "s3_logs_expiration_days" {
  description = "Days before expiring logs"
  type        = number
  default     = 365
}

variable "s3_enable_access_logging" {
  description = "Enable S3 access logging"
  type        = bool
  default     = true
}

# Monitoring Configuration
variable "create_cloudwatch_alarms" {
  description = "Create CloudWatch alarms for monitoring"
  type        = bool
  default     = true
}

variable "alarm_sns_topic_arns" {
  description = "SNS topic ARNs for alarm notifications"
  type        = list(string)
  default     = []
}
