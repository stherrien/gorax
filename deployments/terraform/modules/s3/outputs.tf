# S3 Module Outputs

output "artifacts_bucket_id" {
  description = "ID of the artifacts bucket"
  value       = aws_s3_bucket.artifacts.id
}

output "artifacts_bucket_arn" {
  description = "ARN of the artifacts bucket"
  value       = aws_s3_bucket.artifacts.arn
}

output "artifacts_bucket_domain_name" {
  description = "Domain name of the artifacts bucket"
  value       = aws_s3_bucket.artifacts.bucket_domain_name
}

output "artifacts_bucket_regional_domain_name" {
  description = "Regional domain name of the artifacts bucket"
  value       = aws_s3_bucket.artifacts.bucket_regional_domain_name
}

output "logs_bucket_id" {
  description = "ID of the logs bucket"
  value       = aws_s3_bucket.logs.id
}

output "logs_bucket_arn" {
  description = "ARN of the logs bucket"
  value       = aws_s3_bucket.logs.arn
}

output "logs_bucket_domain_name" {
  description = "Domain name of the logs bucket"
  value       = aws_s3_bucket.logs.bucket_domain_name
}

output "logs_bucket_regional_domain_name" {
  description = "Regional domain name of the logs bucket"
  value       = aws_s3_bucket.logs.bucket_regional_domain_name
}

output "all_bucket_arns" {
  description = "List of all bucket ARNs"
  value = [
    aws_s3_bucket.artifacts.arn,
    aws_s3_bucket.logs.arn
  ]
}
