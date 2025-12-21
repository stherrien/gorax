# SQS Module Outputs

output "main_queue_id" {
  description = "ID (URL) of the main workflow queue"
  value       = aws_sqs_queue.main.id
}

output "main_queue_arn" {
  description = "ARN of the main workflow queue"
  value       = aws_sqs_queue.main.arn
}

output "main_queue_name" {
  description = "Name of the main workflow queue"
  value       = aws_sqs_queue.main.name
}

output "main_dlq_id" {
  description = "ID (URL) of the main dead letter queue"
  value       = aws_sqs_queue.main_dlq.id
}

output "main_dlq_arn" {
  description = "ARN of the main dead letter queue"
  value       = aws_sqs_queue.main_dlq.arn
}

output "main_dlq_name" {
  description = "Name of the main dead letter queue"
  value       = aws_sqs_queue.main_dlq.name
}

output "high_priority_queue_id" {
  description = "ID (URL) of the high priority queue"
  value       = aws_sqs_queue.high_priority.id
}

output "high_priority_queue_arn" {
  description = "ARN of the high priority queue"
  value       = aws_sqs_queue.high_priority.arn
}

output "high_priority_queue_name" {
  description = "Name of the high priority queue"
  value       = aws_sqs_queue.high_priority.name
}

output "high_priority_dlq_id" {
  description = "ID (URL) of the high priority dead letter queue"
  value       = aws_sqs_queue.high_priority_dlq.id
}

output "high_priority_dlq_arn" {
  description = "ARN of the high priority dead letter queue"
  value       = aws_sqs_queue.high_priority_dlq.arn
}

output "high_priority_dlq_name" {
  description = "Name of the high priority dead letter queue"
  value       = aws_sqs_queue.high_priority_dlq.name
}

output "fifo_queue_id" {
  description = "ID (URL) of the FIFO queue"
  value       = var.create_fifo_queue ? aws_sqs_queue.fifo[0].id : null
}

output "fifo_queue_arn" {
  description = "ARN of the FIFO queue"
  value       = var.create_fifo_queue ? aws_sqs_queue.fifo[0].arn : null
}

output "fifo_queue_name" {
  description = "Name of the FIFO queue"
  value       = var.create_fifo_queue ? aws_sqs_queue.fifo[0].name : null
}

output "fifo_dlq_id" {
  description = "ID (URL) of the FIFO dead letter queue"
  value       = var.create_fifo_queue ? aws_sqs_queue.fifo_dlq[0].id : null
}

output "fifo_dlq_arn" {
  description = "ARN of the FIFO dead letter queue"
  value       = var.create_fifo_queue ? aws_sqs_queue.fifo_dlq[0].arn : null
}

output "fifo_dlq_name" {
  description = "Name of the FIFO dead letter queue"
  value       = var.create_fifo_queue ? aws_sqs_queue.fifo_dlq[0].name : null
}

output "all_queue_arns" {
  description = "List of all queue ARNs"
  value = concat(
    [
      aws_sqs_queue.main.arn,
      aws_sqs_queue.main_dlq.arn,
      aws_sqs_queue.high_priority.arn,
      aws_sqs_queue.high_priority_dlq.arn
    ],
    var.create_fifo_queue ? [aws_sqs_queue.fifo[0].arn, aws_sqs_queue.fifo_dlq[0].arn] : []
  )
}
