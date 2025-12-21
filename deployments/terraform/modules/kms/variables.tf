# KMS Module Variables

variable "name_prefix" {
  description = "Prefix for resource names"
  type        = string
}

variable "deletion_window_in_days" {
  description = "Number of days before a KMS key is deleted after destruction"
  type        = number
  default     = 30
  validation {
    condition     = var.deletion_window_in_days >= 7 && var.deletion_window_in_days <= 30
    error_message = "Deletion window must be between 7 and 30 days."
  }
}

variable "tags" {
  description = "Tags to apply to all resources"
  type        = map(string)
  default     = {}
}
