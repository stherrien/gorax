# SQS Module Variables

variable "name_prefix" {
  description = "Prefix for resource names"
  type        = string
}

variable "kms_key_id" {
  description = "KMS key ID for message encryption"
  type        = string
}

variable "kms_data_key_reuse_period_seconds" {
  description = "Length of time in seconds for which Amazon SQS can reuse a data key"
  type        = number
  default     = 300
  validation {
    condition     = var.kms_data_key_reuse_period_seconds >= 60 && var.kms_data_key_reuse_period_seconds <= 86400
    error_message = "KMS data key reuse period must be between 60 and 86400 seconds."
  }
}

variable "allowed_principal_arns" {
  description = "List of IAM principal ARNs allowed to access the queues"
  type        = list(string)
  default     = ["*"]
}

variable "delay_seconds" {
  description = "Seconds to delay message delivery"
  type        = number
  default     = 0
  validation {
    condition     = var.delay_seconds >= 0 && var.delay_seconds <= 900
    error_message = "Delay seconds must be between 0 and 900."
  }
}

variable "max_message_size" {
  description = "Maximum message size in bytes"
  type        = number
  default     = 262144
  validation {
    condition     = var.max_message_size >= 1024 && var.max_message_size <= 262144
    error_message = "Max message size must be between 1024 and 262144 bytes."
  }
}

variable "message_retention_seconds" {
  description = "Number of seconds to retain messages"
  type        = number
  default     = 345600
  validation {
    condition     = var.message_retention_seconds >= 60 && var.message_retention_seconds <= 1209600
    error_message = "Message retention must be between 60 and 1209600 seconds (14 days)."
  }
}

variable "receive_wait_time_seconds" {
  description = "Wait time for ReceiveMessage call (long polling)"
  type        = number
  default     = 20
  validation {
    condition     = var.receive_wait_time_seconds >= 0 && var.receive_wait_time_seconds <= 20
    error_message = "Receive wait time must be between 0 and 20 seconds."
  }
}

variable "visibility_timeout_seconds" {
  description = "Visibility timeout for main queue"
  type        = number
  default     = 300
  validation {
    condition     = var.visibility_timeout_seconds >= 0 && var.visibility_timeout_seconds <= 43200
    error_message = "Visibility timeout must be between 0 and 43200 seconds (12 hours)."
  }
}

variable "high_priority_visibility_timeout_seconds" {
  description = "Visibility timeout for high priority queue"
  type        = number
  default     = 60
  validation {
    condition     = var.high_priority_visibility_timeout_seconds >= 0 && var.high_priority_visibility_timeout_seconds <= 43200
    error_message = "Visibility timeout must be between 0 and 43200 seconds (12 hours)."
  }
}

variable "max_receive_count" {
  description = "Maximum number of receives before sending to DLQ"
  type        = number
  default     = 3
  validation {
    condition     = var.max_receive_count >= 1 && var.max_receive_count <= 1000
    error_message = "Max receive count must be between 1 and 1000."
  }
}

variable "dlq_message_retention_seconds" {
  description = "Number of seconds to retain messages in DLQ"
  type        = number
  default     = 1209600
  validation {
    condition     = var.dlq_message_retention_seconds >= 60 && var.dlq_message_retention_seconds <= 1209600
    error_message = "DLQ message retention must be between 60 and 1209600 seconds (14 days)."
  }
}

variable "create_fifo_queue" {
  description = "Create FIFO queue for ordered message processing"
  type        = bool
  default     = false
}

variable "create_cloudwatch_alarms" {
  description = "Create CloudWatch alarms for queue monitoring"
  type        = bool
  default     = true
}

variable "age_alarm_threshold" {
  description = "Threshold in seconds for oldest message age alarm"
  type        = number
  default     = 600
}

variable "high_priority_age_alarm_threshold" {
  description = "Threshold in seconds for high priority queue oldest message age alarm"
  type        = number
  default     = 120
}

variable "depth_alarm_threshold" {
  description = "Threshold for queue depth alarm"
  type        = number
  default     = 1000
}

variable "dlq_depth_alarm_threshold" {
  description = "Threshold for DLQ depth alarm"
  type        = number
  default     = 10
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
