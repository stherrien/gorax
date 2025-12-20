# KMS Module Outputs

output "database_key_id" {
  description = "ID of the database encryption key"
  value       = aws_kms_key.database.id
}

output "database_key_arn" {
  description = "ARN of the database encryption key"
  value       = aws_kms_key.database.arn
}

output "s3_key_id" {
  description = "ID of the S3 encryption key"
  value       = aws_kms_key.s3.id
}

output "s3_key_arn" {
  description = "ARN of the S3 encryption key"
  value       = aws_kms_key.s3.arn
}

output "secrets_key_id" {
  description = "ID of the Secrets Manager encryption key"
  value       = aws_kms_key.secrets.id
}

output "secrets_key_arn" {
  description = "ARN of the Secrets Manager encryption key"
  value       = aws_kms_key.secrets.arn
}

output "elasticache_key_id" {
  description = "ID of the ElastiCache encryption key"
  value       = aws_kms_key.elasticache.id
}

output "elasticache_key_arn" {
  description = "ARN of the ElastiCache encryption key"
  value       = aws_kms_key.elasticache.arn
}

output "sqs_key_id" {
  description = "ID of the SQS encryption key"
  value       = aws_kms_key.sqs.id
}

output "sqs_key_arn" {
  description = "ARN of the SQS encryption key"
  value       = aws_kms_key.sqs.arn
}

output "eks_key_id" {
  description = "ID of the EKS encryption key"
  value       = aws_kms_key.eks.id
}

output "eks_key_arn" {
  description = "ARN of the EKS encryption key"
  value       = aws_kms_key.eks.arn
}
