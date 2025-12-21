# Aurora Module Variables

variable "name_prefix" {
  description = "Prefix for resource names"
  type        = string
}

variable "vpc_id" {
  description = "VPC ID where Aurora cluster will be created"
  type        = string
}

variable "subnet_ids" {
  description = "List of subnet IDs for the DB subnet group"
  type        = list(string)
}

variable "allowed_security_group_ids" {
  description = "List of security group IDs allowed to access the database"
  type        = list(string)
}

variable "availability_zones" {
  description = "List of availability zones for the cluster"
  type        = list(string)
}

variable "database_name" {
  description = "Name of the default database to create"
  type        = string
  default     = "gorax"
}

variable "master_username" {
  description = "Master username for the database"
  type        = string
  default     = "gorax_admin"
}

variable "engine_version" {
  description = "Aurora PostgreSQL engine version"
  type        = string
  default     = "15.4"
}

variable "engine_family" {
  description = "Aurora PostgreSQL engine family for parameter groups"
  type        = string
  default     = "aurora-postgresql15"
}

variable "engine_mode" {
  description = "Engine mode (provisioned or serverless)"
  type        = string
  default     = "provisioned"
  validation {
    condition     = contains(["provisioned", "serverless"], var.engine_mode)
    error_message = "Engine mode must be either 'provisioned' or 'serverless'."
  }
}

variable "enable_serverless_v2" {
  description = "Enable Serverless v2 scaling (only for provisioned mode)"
  type        = bool
  default     = true
}

variable "serverless_min_capacity" {
  description = "Minimum capacity units for Serverless v2"
  type        = number
  default     = 0.5
}

variable "serverless_max_capacity" {
  description = "Maximum capacity units for Serverless v2"
  type        = number
  default     = 2
}

variable "instance_class" {
  description = "Instance class for provisioned instances (if not using Serverless v2)"
  type        = string
  default     = "db.t4g.medium"
}

variable "instance_count" {
  description = "Number of database instances to create"
  type        = number
  default     = 2
  validation {
    condition     = var.instance_count >= 1 && var.instance_count <= 15
    error_message = "Instance count must be between 1 and 15."
  }
}

variable "backup_retention_period" {
  description = "Number of days to retain backups"
  type        = number
  default     = 7
  validation {
    condition     = var.backup_retention_period >= 1 && var.backup_retention_period <= 35
    error_message = "Backup retention period must be between 1 and 35 days."
  }
}

variable "preferred_backup_window" {
  description = "Preferred backup window (UTC)"
  type        = string
  default     = "03:00-04:00"
}

variable "preferred_maintenance_window" {
  description = "Preferred maintenance window (UTC)"
  type        = string
  default     = "sun:04:00-sun:05:00"
}

variable "skip_final_snapshot" {
  description = "Skip final snapshot when destroying the cluster"
  type        = bool
  default     = false
}

variable "deletion_protection" {
  description = "Enable deletion protection"
  type        = bool
  default     = true
}

variable "kms_key_arn" {
  description = "KMS key ARN for encryption at rest"
  type        = string
}

variable "secrets_kms_key_arn" {
  description = "KMS key ARN for Secrets Manager"
  type        = string
}

variable "enabled_cloudwatch_logs_exports" {
  description = "List of log types to export to CloudWatch"
  type        = list(string)
  default     = ["postgresql"]
}

variable "monitoring_interval" {
  description = "Enhanced monitoring interval in seconds (0, 1, 5, 10, 15, 30, 60)"
  type        = number
  default     = 60
  validation {
    condition     = contains([0, 1, 5, 10, 15, 30, 60], var.monitoring_interval)
    error_message = "Monitoring interval must be 0, 1, 5, 10, 15, 30, or 60."
  }
}

variable "enable_performance_insights" {
  description = "Enable Performance Insights"
  type        = bool
  default     = true
}

variable "performance_insights_retention_period" {
  description = "Performance Insights retention period in days"
  type        = number
  default     = 7
  validation {
    condition     = contains([7, 731], var.performance_insights_retention_period)
    error_message = "Performance Insights retention must be 7 or 731 days."
  }
}

variable "auto_minor_version_upgrade" {
  description = "Enable automatic minor version upgrades"
  type        = bool
  default     = true
}

variable "max_connections" {
  description = "Maximum number of database connections"
  type        = string
  default     = "1000"
}

variable "secret_recovery_window_days" {
  description = "Number of days to recover deleted secrets"
  type        = number
  default     = 7
  validation {
    condition     = var.secret_recovery_window_days >= 7 && var.secret_recovery_window_days <= 30
    error_message = "Secret recovery window must be between 7 and 30 days."
  }
}

variable "create_cloudwatch_alarms" {
  description = "Create CloudWatch alarms for database monitoring"
  type        = bool
  default     = true
}

variable "cpu_alarm_threshold" {
  description = "CPU utilization threshold for CloudWatch alarm"
  type        = number
  default     = 80
}

variable "alarm_actions" {
  description = "List of ARNs to notify when alarm triggers"
  type        = list(string)
  default     = []
}

variable "tags" {
  description = "Tags to apply to all resources"
  type        = map(string)
  default     = {}
}
