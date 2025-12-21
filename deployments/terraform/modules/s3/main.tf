# S3 Module - Storage Buckets for Gorax
# Creates S3 buckets with encryption, versioning, and lifecycle policies

data "aws_caller_identity" "current" {}

# Artifacts Bucket (workflow definitions backup)
resource "aws_s3_bucket" "artifacts" {
  bucket = "${var.name_prefix}-artifacts-${data.aws_caller_identity.current.account_id}"

  tags = merge(
    var.tags,
    {
      Name = "${var.name_prefix}-artifacts"
      Type = "artifacts"
    }
  )
}

resource "aws_s3_bucket_versioning" "artifacts" {
  bucket = aws_s3_bucket.artifacts.id

  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "artifacts" {
  bucket = aws_s3_bucket.artifacts.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm     = "aws:kms"
      kms_master_key_id = var.kms_key_arn
    }
    bucket_key_enabled = true
  }
}

resource "aws_s3_bucket_public_access_block" "artifacts" {
  bucket = aws_s3_bucket.artifacts.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_bucket_lifecycle_configuration" "artifacts" {
  bucket = aws_s3_bucket.artifacts.id

  rule {
    id     = "transition-to-ia"
    status = "Enabled"

    transition {
      days          = var.artifacts_transition_to_ia_days
      storage_class = "STANDARD_IA"
    }

    transition {
      days          = var.artifacts_transition_to_glacier_days
      storage_class = "GLACIER"
    }

    noncurrent_version_transition {
      noncurrent_days = 30
      storage_class   = "STANDARD_IA"
    }

    noncurrent_version_transition {
      noncurrent_days = 90
      storage_class   = "GLACIER"
    }

    noncurrent_version_expiration {
      noncurrent_days = var.artifacts_noncurrent_version_expiration_days
    }
  }
}

resource "aws_s3_bucket_logging" "artifacts" {
  count  = var.enable_access_logging ? 1 : 0
  bucket = aws_s3_bucket.artifacts.id

  target_bucket = aws_s3_bucket.logs.id
  target_prefix = "artifacts-access-logs/"
}

# Logs Bucket (execution logs)
resource "aws_s3_bucket" "logs" {
  bucket = "${var.name_prefix}-logs-${data.aws_caller_identity.current.account_id}"

  tags = merge(
    var.tags,
    {
      Name = "${var.name_prefix}-logs"
      Type = "logs"
    }
  )
}

resource "aws_s3_bucket_versioning" "logs" {
  bucket = aws_s3_bucket.logs.id

  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "logs" {
  bucket = aws_s3_bucket.logs.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm     = "aws:kms"
      kms_master_key_id = var.kms_key_arn
    }
    bucket_key_enabled = true
  }
}

resource "aws_s3_bucket_public_access_block" "logs" {
  bucket = aws_s3_bucket.logs.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_bucket_lifecycle_configuration" "logs" {
  bucket = aws_s3_bucket.logs.id

  rule {
    id     = "transition-and-expire"
    status = "Enabled"

    transition {
      days          = var.logs_transition_to_ia_days
      storage_class = "STANDARD_IA"
    }

    transition {
      days          = var.logs_transition_to_glacier_days
      storage_class = "GLACIER"
    }

    expiration {
      days = var.logs_expiration_days
    }

    noncurrent_version_expiration {
      noncurrent_days = var.logs_noncurrent_version_expiration_days
    }
  }
}

# Allow S3 to write access logs to logs bucket
resource "aws_s3_bucket_policy" "logs" {
  bucket = aws_s3_bucket.logs.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "S3ServerAccessLogsPolicy"
        Effect = "Allow"
        Principal = {
          Service = "logging.s3.amazonaws.com"
        }
        Action   = "s3:PutObject"
        Resource = "${aws_s3_bucket.logs.arn}/*"
        Condition = {
          StringEquals = {
            "aws:SourceAccount" = data.aws_caller_identity.current.account_id
          }
        }
      }
    ]
  })
}

# Bucket Policies for Application Access
resource "aws_s3_bucket_policy" "artifacts" {
  bucket = aws_s3_bucket.artifacts.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "DenyUnencryptedObjectUploads"
        Effect = "Deny"
        Principal = "*"
        Action = "s3:PutObject"
        Resource = "${aws_s3_bucket.artifacts.arn}/*"
        Condition = {
          StringNotEquals = {
            "s3:x-amz-server-side-encryption" = "aws:kms"
          }
        }
      },
      {
        Sid    = "DenyInsecureTransport"
        Effect = "Deny"
        Principal = "*"
        Action = "s3:*"
        Resource = [
          aws_s3_bucket.artifacts.arn,
          "${aws_s3_bucket.artifacts.arn}/*"
        ]
        Condition = {
          Bool = {
            "aws:SecureTransport" = "false"
          }
        }
      },
      {
        Sid    = "AllowApplicationAccess"
        Effect = "Allow"
        Principal = {
          AWS = var.allowed_principal_arns
        }
        Action = [
          "s3:GetObject",
          "s3:PutObject",
          "s3:DeleteObject",
          "s3:ListBucket"
        ]
        Resource = [
          aws_s3_bucket.artifacts.arn,
          "${aws_s3_bucket.artifacts.arn}/*"
        ]
      }
    ]
  })
}

# Bucket Notifications (optional)
resource "aws_s3_bucket_notification" "artifacts" {
  count  = var.enable_event_notifications ? 1 : 0
  bucket = aws_s3_bucket.artifacts.id

  dynamic "topic" {
    for_each = var.notification_topic_arns
    content {
      topic_arn = topic.value
      events = [
        "s3:ObjectCreated:*",
        "s3:ObjectRemoved:*"
      ]
    }
  }
}

# CloudWatch Alarms
resource "aws_cloudwatch_metric_alarm" "artifacts_bucket_size" {
  count               = var.create_cloudwatch_alarms ? 1 : 0
  alarm_name          = "${var.name_prefix}-artifacts-bucket-size"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "1"
  metric_name         = "BucketSizeBytes"
  namespace           = "AWS/S3"
  period              = "86400"
  statistic           = "Average"
  threshold           = var.bucket_size_alarm_threshold
  alarm_description   = "This metric monitors artifacts bucket size"
  alarm_actions       = var.alarm_actions

  dimensions = {
    BucketName  = aws_s3_bucket.artifacts.id
    StorageType = "StandardStorage"
  }

  tags = var.tags
}

resource "aws_cloudwatch_metric_alarm" "logs_bucket_size" {
  count               = var.create_cloudwatch_alarms ? 1 : 0
  alarm_name          = "${var.name_prefix}-logs-bucket-size"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "1"
  metric_name         = "BucketSizeBytes"
  namespace           = "AWS/S3"
  period              = "86400"
  statistic           = "Average"
  threshold           = var.bucket_size_alarm_threshold
  alarm_description   = "This metric monitors logs bucket size"
  alarm_actions       = var.alarm_actions

  dimensions = {
    BucketName  = aws_s3_bucket.logs.id
    StorageType = "StandardStorage"
  }

  tags = var.tags
}

# Replication Configuration (optional for cross-region backup)
resource "aws_s3_bucket_replication_configuration" "artifacts" {
  count = var.enable_cross_region_replication ? 1 : 0

  depends_on = [
    aws_s3_bucket_versioning.artifacts
  ]

  role   = var.replication_role_arn
  bucket = aws_s3_bucket.artifacts.id

  rule {
    id     = "replicate-all"
    status = "Enabled"

    destination {
      bucket        = var.replication_destination_bucket_arn
      storage_class = "STANDARD_IA"

      encryption_configuration {
        replica_kms_key_id = var.replication_kms_key_arn
      }
    }
  }
}
