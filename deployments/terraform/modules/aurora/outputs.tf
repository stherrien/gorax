# Aurora Module Outputs

output "cluster_id" {
  description = "ID of the Aurora cluster"
  value       = aws_rds_cluster.main.id
}

output "cluster_arn" {
  description = "ARN of the Aurora cluster"
  value       = aws_rds_cluster.main.arn
}

output "cluster_endpoint" {
  description = "Writer endpoint for the Aurora cluster"
  value       = aws_rds_cluster.main.endpoint
}

output "cluster_reader_endpoint" {
  description = "Reader endpoint for the Aurora cluster"
  value       = aws_rds_cluster.main.reader_endpoint
}

output "cluster_port" {
  description = "Port on which the database accepts connections"
  value       = aws_rds_cluster.main.port
}

output "database_name" {
  description = "Name of the default database"
  value       = aws_rds_cluster.main.database_name
}

output "master_username" {
  description = "Master username for the database"
  value       = aws_rds_cluster.main.master_username
  sensitive   = true
}

output "security_group_id" {
  description = "ID of the security group for Aurora cluster"
  value       = aws_security_group.aurora.id
}

output "secret_arn" {
  description = "ARN of the Secrets Manager secret containing database credentials"
  value       = aws_secretsmanager_secret.db_credentials.arn
}

output "secret_name" {
  description = "Name of the Secrets Manager secret containing database credentials"
  value       = aws_secretsmanager_secret.db_credentials.name
}

output "instance_ids" {
  description = "List of instance IDs"
  value       = aws_rds_cluster_instance.main[*].id
}

output "instance_endpoints" {
  description = "List of instance endpoints"
  value       = aws_rds_cluster_instance.main[*].endpoint
}

output "cluster_resource_id" {
  description = "Resource ID of the Aurora cluster"
  value       = aws_rds_cluster.main.cluster_resource_id
}
