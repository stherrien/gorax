# S3 Module Variables

variable "name_prefix" {
  description = "Prefix for resource names"
  type        = string
}

variable "kms_key_arn" {
  description = "KMS key ARN for bucket encryption"
  type        = string
}

variable "allowed_principal_arns" {
  description = "List of IAM principal ARNs allowed to access the buckets"
  type        = list(string)
  default     = ["*"]
}

# Artifacts Bucket Lifecycle
variable "artifacts_transition_to_ia_days" {
  description = "Number of days before transitioning artifacts to Infrequent Access"
  type        = number
  default     = 90
}

variable "artifacts_transition_to_glacier_days" {
  description = "Number of days before transitioning artifacts to Glacier"
  type        = number
  default     = 180
}

variable "artifacts_noncurrent_version_expiration_days" {
  description = "Number of days to keep non-current versions of artifacts"
  type        = number
  default     = 365
}

# Logs Bucket Lifecycle
variable "logs_transition_to_ia_days" {
  description = "Number of days before transitioning logs to Infrequent Access"
  type        = number
  default     = 30
}

variable "logs_transition_to_glacier_days" {
  description = "Number of days before transitioning logs to Glacier"
  type        = number
  default     = 90
}

variable "logs_expiration_days" {
  description = "Number of days before expiring logs"
  type        = number
  default     = 365
}

variable "logs_noncurrent_version_expiration_days" {
  description = "Number of days to keep non-current versions of logs"
  type        = number
  default     = 30
}

# Access Logging
variable "enable_access_logging" {
  description = "Enable S3 access logging"
  type        = bool
  default     = true
}

# Event Notifications
variable "enable_event_notifications" {
  description = "Enable S3 event notifications"
  type        = bool
  default     = false
}

variable "notification_topic_arns" {
  description = "List of SNS topic ARNs for bucket notifications"
  type        = list(string)
  default     = []
}

# Cross-Region Replication
variable "enable_cross_region_replication" {
  description = "Enable cross-region replication for disaster recovery"
  type        = bool
  default     = false
}

variable "replication_role_arn" {
  description = "IAM role ARN for S3 replication"
  type        = string
  default     = null
}

variable "replication_destination_bucket_arn" {
  description = "Destination bucket ARN for replication"
  type        = string
  default     = null
}

variable "replication_kms_key_arn" {
  description = "KMS key ARN for replication destination encryption"
  type        = string
  default     = null
}

# CloudWatch Alarms
variable "create_cloudwatch_alarms" {
  description = "Create CloudWatch alarms for bucket monitoring"
  type        = bool
  default     = true
}

variable "bucket_size_alarm_threshold" {
  description = "Threshold in bytes for bucket size alarm"
  type        = number
  default     = 107374182400
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
