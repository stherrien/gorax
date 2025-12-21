# ElastiCache Module Outputs

output "replication_group_id" {
  description = "ID of the ElastiCache replication group"
  value       = aws_elasticache_replication_group.main.id
}

output "replication_group_arn" {
  description = "ARN of the ElastiCache replication group"
  value       = aws_elasticache_replication_group.main.arn
}

output "primary_endpoint_address" {
  description = "Primary endpoint address for Redis cluster"
  value       = aws_elasticache_replication_group.main.primary_endpoint_address
}

output "reader_endpoint_address" {
  description = "Reader endpoint address for Redis cluster"
  value       = aws_elasticache_replication_group.main.reader_endpoint_address
}

output "port" {
  description = "Port number on which Redis accepts connections"
  value       = aws_elasticache_replication_group.main.port
}

output "security_group_id" {
  description = "ID of the security group for Redis cluster"
  value       = aws_security_group.redis.id
}

output "secret_arn" {
  description = "ARN of the Secrets Manager secret containing Redis credentials"
  value       = var.transit_encryption_enabled ? aws_secretsmanager_secret.redis_credentials[0].arn : null
}

output "secret_name" {
  description = "Name of the Secrets Manager secret containing Redis credentials"
  value       = var.transit_encryption_enabled ? aws_secretsmanager_secret.redis_credentials[0].name : null
}

output "member_clusters" {
  description = "List of member cluster IDs"
  value       = aws_elasticache_replication_group.main.member_clusters
}

output "configuration_endpoint_address" {
  description = "Configuration endpoint address (for cluster mode)"
  value       = aws_elasticache_replication_group.main.configuration_endpoint_address
}
