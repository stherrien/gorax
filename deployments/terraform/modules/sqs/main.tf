# SQS Module - Message Queues for Gorax
# Creates SQS queues with dead letter queues and encryption

# Dead Letter Queue for Main Queue
resource "aws_sqs_queue" "main_dlq" {
  name                       = "${var.name_prefix}-workflow-dlq"
  message_retention_seconds  = var.dlq_message_retention_seconds
  receive_wait_time_seconds  = var.receive_wait_time_seconds
  kms_master_key_id          = var.kms_key_id
  kms_data_key_reuse_period_seconds = var.kms_data_key_reuse_period_seconds

  tags = merge(
    var.tags,
    {
      Name = "${var.name_prefix}-workflow-dlq"
      Type = "dead-letter-queue"
    }
  )
}

# Main Workflow Queue
resource "aws_sqs_queue" "main" {
  name                       = "${var.name_prefix}-workflow-queue"
  delay_seconds              = var.delay_seconds
  max_message_size           = var.max_message_size
  message_retention_seconds  = var.message_retention_seconds
  receive_wait_time_seconds  = var.receive_wait_time_seconds
  visibility_timeout_seconds = var.visibility_timeout_seconds
  kms_master_key_id          = var.kms_key_id
  kms_data_key_reuse_period_seconds = var.kms_data_key_reuse_period_seconds

  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.main_dlq.arn
    maxReceiveCount     = var.max_receive_count
  })

  tags = merge(
    var.tags,
    {
      Name = "${var.name_prefix}-workflow-queue"
      Type = "main-queue"
    }
  )
}

# Dead Letter Queue for High Priority Queue
resource "aws_sqs_queue" "high_priority_dlq" {
  name                       = "${var.name_prefix}-workflow-high-priority-dlq"
  message_retention_seconds  = var.dlq_message_retention_seconds
  receive_wait_time_seconds  = var.receive_wait_time_seconds
  kms_master_key_id          = var.kms_key_id
  kms_data_key_reuse_period_seconds = var.kms_data_key_reuse_period_seconds

  tags = merge(
    var.tags,
    {
      Name = "${var.name_prefix}-workflow-high-priority-dlq"
      Type = "dead-letter-queue"
    }
  )
}

# High Priority Queue
resource "aws_sqs_queue" "high_priority" {
  name                       = "${var.name_prefix}-workflow-high-priority-queue"
  delay_seconds              = var.delay_seconds
  max_message_size           = var.max_message_size
  message_retention_seconds  = var.message_retention_seconds
  receive_wait_time_seconds  = var.receive_wait_time_seconds
  visibility_timeout_seconds = var.high_priority_visibility_timeout_seconds
  kms_master_key_id          = var.kms_key_id
  kms_data_key_reuse_period_seconds = var.kms_data_key_reuse_period_seconds

  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.high_priority_dlq.arn
    maxReceiveCount     = var.max_receive_count
  })

  tags = merge(
    var.tags,
    {
      Name = "${var.name_prefix}-workflow-high-priority-queue"
      Type = "high-priority-queue"
    }
  )
}

# FIFO Dead Letter Queue (optional)
resource "aws_sqs_queue" "fifo_dlq" {
  count                      = var.create_fifo_queue ? 1 : 0
  name                       = "${var.name_prefix}-workflow-fifo.fifo"
  fifo_queue                 = true
  content_based_deduplication = true
  message_retention_seconds  = var.dlq_message_retention_seconds
  receive_wait_time_seconds  = var.receive_wait_time_seconds
  kms_master_key_id          = var.kms_key_id
  kms_data_key_reuse_period_seconds = var.kms_data_key_reuse_period_seconds

  tags = merge(
    var.tags,
    {
      Name = "${var.name_prefix}-workflow-fifo-dlq"
      Type = "dead-letter-queue"
    }
  )
}

# FIFO Queue (optional)
resource "aws_sqs_queue" "fifo" {
  count                      = var.create_fifo_queue ? 1 : 0
  name                       = "${var.name_prefix}-workflow-fifo.fifo"
  fifo_queue                 = true
  content_based_deduplication = true
  delay_seconds              = var.delay_seconds
  max_message_size           = var.max_message_size
  message_retention_seconds  = var.message_retention_seconds
  receive_wait_time_seconds  = var.receive_wait_time_seconds
  visibility_timeout_seconds = var.visibility_timeout_seconds
  kms_master_key_id          = var.kms_key_id
  kms_data_key_reuse_period_seconds = var.kms_data_key_reuse_period_seconds

  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.fifo_dlq[0].arn
    maxReceiveCount     = var.max_receive_count
  })

  tags = merge(
    var.tags,
    {
      Name = "${var.name_prefix}-workflow-fifo-queue"
      Type = "fifo-queue"
    }
  )
}

# Queue Policies
resource "aws_sqs_queue_policy" "main" {
  queue_url = aws_sqs_queue.main.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "AllowSendMessage"
        Effect = "Allow"
        Principal = {
          AWS = var.allowed_principal_arns
        }
        Action = [
          "sqs:SendMessage",
          "sqs:SendMessageBatch"
        ]
        Resource = aws_sqs_queue.main.arn
      },
      {
        Sid    = "AllowReceiveMessage"
        Effect = "Allow"
        Principal = {
          AWS = var.allowed_principal_arns
        }
        Action = [
          "sqs:ReceiveMessage",
          "sqs:DeleteMessage",
          "sqs:ChangeMessageVisibility",
          "sqs:GetQueueAttributes"
        ]
        Resource = aws_sqs_queue.main.arn
      }
    ]
  })
}

resource "aws_sqs_queue_policy" "high_priority" {
  queue_url = aws_sqs_queue.high_priority.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "AllowSendMessage"
        Effect = "Allow"
        Principal = {
          AWS = var.allowed_principal_arns
        }
        Action = [
          "sqs:SendMessage",
          "sqs:SendMessageBatch"
        ]
        Resource = aws_sqs_queue.high_priority.arn
      },
      {
        Sid    = "AllowReceiveMessage"
        Effect = "Allow"
        Principal = {
          AWS = var.allowed_principal_arns
        }
        Action = [
          "sqs:ReceiveMessage",
          "sqs:DeleteMessage",
          "sqs:ChangeMessageVisibility",
          "sqs:GetQueueAttributes"
        ]
        Resource = aws_sqs_queue.high_priority.arn
      }
    ]
  })
}

# CloudWatch Alarms for Main Queue
resource "aws_cloudwatch_metric_alarm" "main_queue_age" {
  count               = var.create_cloudwatch_alarms ? 1 : 0
  alarm_name          = "${var.name_prefix}-workflow-queue-age"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "2"
  metric_name         = "ApproximateAgeOfOldestMessage"
  namespace           = "AWS/SQS"
  period              = "300"
  statistic           = "Maximum"
  threshold           = var.age_alarm_threshold
  alarm_description   = "This metric monitors the age of oldest message in queue"
  alarm_actions       = var.alarm_actions

  dimensions = {
    QueueName = aws_sqs_queue.main.name
  }

  tags = var.tags
}

resource "aws_cloudwatch_metric_alarm" "main_queue_depth" {
  count               = var.create_cloudwatch_alarms ? 1 : 0
  alarm_name          = "${var.name_prefix}-workflow-queue-depth"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "2"
  metric_name         = "ApproximateNumberOfMessagesVisible"
  namespace           = "AWS/SQS"
  period              = "300"
  statistic           = "Average"
  threshold           = var.depth_alarm_threshold
  alarm_description   = "This metric monitors queue depth"
  alarm_actions       = var.alarm_actions

  dimensions = {
    QueueName = aws_sqs_queue.main.name
  }

  tags = var.tags
}

resource "aws_cloudwatch_metric_alarm" "main_dlq_depth" {
  count               = var.create_cloudwatch_alarms ? 1 : 0
  alarm_name          = "${var.name_prefix}-workflow-dlq-depth"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "1"
  metric_name         = "ApproximateNumberOfMessagesVisible"
  namespace           = "AWS/SQS"
  period              = "300"
  statistic           = "Average"
  threshold           = var.dlq_depth_alarm_threshold
  alarm_description   = "This metric monitors dead letter queue depth"
  alarm_actions       = var.alarm_actions

  dimensions = {
    QueueName = aws_sqs_queue.main_dlq.name
  }

  tags = var.tags
}

# CloudWatch Alarms for High Priority Queue
resource "aws_cloudwatch_metric_alarm" "high_priority_queue_age" {
  count               = var.create_cloudwatch_alarms ? 1 : 0
  alarm_name          = "${var.name_prefix}-workflow-high-priority-queue-age"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "2"
  metric_name         = "ApproximateAgeOfOldestMessage"
  namespace           = "AWS/SQS"
  period              = "60"
  statistic           = "Maximum"
  threshold           = var.high_priority_age_alarm_threshold
  alarm_description   = "This metric monitors the age of oldest message in high priority queue"
  alarm_actions       = var.alarm_actions

  dimensions = {
    QueueName = aws_sqs_queue.high_priority.name
  }

  tags = var.tags
}

resource "aws_cloudwatch_metric_alarm" "high_priority_dlq_depth" {
  count               = var.create_cloudwatch_alarms ? 1 : 0
  alarm_name          = "${var.name_prefix}-workflow-high-priority-dlq-depth"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "1"
  metric_name         = "ApproximateNumberOfMessagesVisible"
  namespace           = "AWS/SQS"
  period              = "300"
  statistic           = "Average"
  threshold           = var.dlq_depth_alarm_threshold
  alarm_description   = "This metric monitors high priority dead letter queue depth"
  alarm_actions       = var.alarm_actions

  dimensions = {
    QueueName = aws_sqs_queue.high_priority_dlq.name
  }

  tags = var.tags
}
