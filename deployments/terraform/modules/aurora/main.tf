# Aurora Module - PostgreSQL Database for Gorax
# Creates Aurora PostgreSQL Serverless v2 cluster

resource "random_password" "master" {
  length  = 32
  special = true
}

# DB Subnet Group
resource "aws_db_subnet_group" "main" {
  name       = "${var.name_prefix}-db-subnet-group"
  subnet_ids = var.subnet_ids

  tags = merge(
    var.tags,
    {
      Name = "${var.name_prefix}-db-subnet-group"
    }
  )
}

# Security Group
resource "aws_security_group" "aurora" {
  name_description = "${var.name_prefix}-aurora-sg"
  description      = "Security group for Aurora PostgreSQL cluster"
  vpc_id           = var.vpc_id

  ingress {
    from_port       = 5432
    to_port         = 5432
    protocol        = "tcp"
    security_groups = var.allowed_security_group_ids
    description     = "PostgreSQL access from EKS"
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
    description = "Allow all outbound"
  }

  tags = merge(
    var.tags,
    {
      Name = "${var.name_prefix}-aurora-sg"
    }
  )
}

# Parameter Group
resource "aws_rds_cluster_parameter_group" "main" {
  name        = "${var.name_prefix}-aurora-pg-cluster-params"
  family      = var.engine_family
  description = "Cluster parameter group for ${var.name_prefix}"

  # Performance and optimization parameters
  parameter {
    name  = "shared_preload_libraries"
    value = "pg_stat_statements"
  }

  parameter {
    name  = "log_statement"
    value = "ddl"
  }

  parameter {
    name  = "log_min_duration_statement"
    value = "1000"
  }

  parameter {
    name  = "log_connections"
    value = "1"
  }

  parameter {
    name  = "log_disconnections"
    value = "1"
  }

  parameter {
    name  = "log_lock_waits"
    value = "1"
  }

  parameter {
    name  = "log_temp_files"
    value = "0"
  }

  parameter {
    name  = "max_connections"
    value = var.max_connections
  }

  tags = var.tags
}

# DB Parameter Group
resource "aws_db_parameter_group" "main" {
  name        = "${var.name_prefix}-aurora-pg-instance-params"
  family      = var.engine_family
  description = "Instance parameter group for ${var.name_prefix}"

  parameter {
    name  = "log_rotation_age"
    value = "1440"
  }

  parameter {
    name  = "log_rotation_size"
    value = "102400"
  }

  tags = var.tags
}

# Aurora Cluster
resource "aws_rds_cluster" "main" {
  cluster_identifier              = "${var.name_prefix}-aurora-cluster"
  engine                          = "aurora-postgresql"
  engine_mode                     = var.engine_mode
  engine_version                  = var.engine_version
  database_name                   = var.database_name
  master_username                 = var.master_username
  master_password                 = random_password.master.result
  db_subnet_group_name            = aws_db_subnet_group.main.name
  vpc_security_group_ids          = [aws_security_group.aurora.id]
  db_cluster_parameter_group_name = aws_rds_cluster_parameter_group.main.name
  port                            = 5432

  # Backup configuration
  backup_retention_period      = var.backup_retention_period
  preferred_backup_window      = var.preferred_backup_window
  preferred_maintenance_window = var.preferred_maintenance_window
  copy_tags_to_snapshot        = true
  skip_final_snapshot          = var.skip_final_snapshot
  final_snapshot_identifier    = var.skip_final_snapshot ? null : "${var.name_prefix}-final-snapshot-${formatdate("YYYY-MM-DD-hhmm", timestamp())}"

  # Encryption
  storage_encrypted = true
  kms_key_id        = var.kms_key_arn

  # Enhanced monitoring
  enabled_cloudwatch_logs_exports = var.enabled_cloudwatch_logs_exports

  # High availability
  availability_zones = var.availability_zones

  # Serverless v2 scaling
  dynamic "serverlessv2_scaling_configuration" {
    for_each = var.engine_mode == "provisioned" && var.enable_serverless_v2 ? [1] : []
    content {
      min_capacity = var.serverless_min_capacity
      max_capacity = var.serverless_max_capacity
    }
  }

  # Deletion protection
  deletion_protection = var.deletion_protection

  tags = merge(
    var.tags,
    {
      Name = "${var.name_prefix}-aurora-cluster"
    }
  )
}

# Aurora Cluster Instances
resource "aws_rds_cluster_instance" "main" {
  count = var.instance_count

  identifier              = "${var.name_prefix}-aurora-instance-${count.index + 1}"
  cluster_identifier      = aws_rds_cluster.main.id
  instance_class          = var.enable_serverless_v2 ? "db.serverless" : var.instance_class
  engine                  = aws_rds_cluster.main.engine
  engine_version          = aws_rds_cluster.main.engine_version
  db_parameter_group_name = aws_db_parameter_group.main.name
  publicly_accessible     = false
  auto_minor_version_upgrade = var.auto_minor_version_upgrade

  # Performance Insights
  performance_insights_enabled    = var.enable_performance_insights
  performance_insights_kms_key_id = var.enable_performance_insights ? var.kms_key_arn : null
  performance_insights_retention_period = var.enable_performance_insights ? var.performance_insights_retention_period : null

  # Monitoring
  monitoring_interval = var.monitoring_interval
  monitoring_role_arn = var.monitoring_interval > 0 ? aws_iam_role.rds_monitoring[0].arn : null

  tags = merge(
    var.tags,
    {
      Name = "${var.name_prefix}-aurora-instance-${count.index + 1}"
    }
  )
}

# IAM Role for Enhanced Monitoring
resource "aws_iam_role" "rds_monitoring" {
  count = var.monitoring_interval > 0 ? 1 : 0
  name  = "${var.name_prefix}-rds-monitoring-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "monitoring.rds.amazonaws.com"
        }
        Action = "sts:AssumeRole"
      }
    ]
  })

  tags = var.tags
}

resource "aws_iam_role_policy_attachment" "rds_monitoring" {
  count      = var.monitoring_interval > 0 ? 1 : 0
  role       = aws_iam_role.rds_monitoring[0].name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonRDSEnhancedMonitoringRole"
}

# Secrets Manager Secret for Database Credentials
resource "aws_secretsmanager_secret" "db_credentials" {
  name_prefix             = "${var.name_prefix}-db-credentials-"
  description             = "Database credentials for ${var.name_prefix} Aurora cluster"
  kms_key_id              = var.secrets_kms_key_arn
  recovery_window_in_days = var.secret_recovery_window_days

  tags = var.tags
}

resource "aws_secretsmanager_secret_version" "db_credentials" {
  secret_id = aws_secretsmanager_secret.db_credentials.id
  secret_string = jsonencode({
    username = aws_rds_cluster.main.master_username
    password = random_password.master.result
    engine   = aws_rds_cluster.main.engine
    host     = aws_rds_cluster.main.endpoint
    port     = aws_rds_cluster.main.port
    dbname   = aws_rds_cluster.main.database_name
    endpoint = aws_rds_cluster.main.endpoint
    reader_endpoint = aws_rds_cluster.main.reader_endpoint
  })
}

# CloudWatch Alarms
resource "aws_cloudwatch_metric_alarm" "database_cpu" {
  count               = var.create_cloudwatch_alarms ? var.instance_count : 0
  alarm_name          = "${var.name_prefix}-aurora-cpu-${count.index + 1}"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "2"
  metric_name         = "CPUUtilization"
  namespace           = "AWS/RDS"
  period              = "300"
  statistic           = "Average"
  threshold           = var.cpu_alarm_threshold
  alarm_description   = "This metric monitors Aurora instance CPU utilization"
  alarm_actions       = var.alarm_actions

  dimensions = {
    DBInstanceIdentifier = aws_rds_cluster_instance.main[count.index].id
  }

  tags = var.tags
}

resource "aws_cloudwatch_metric_alarm" "database_connections" {
  count               = var.create_cloudwatch_alarms ? 1 : 0
  alarm_name          = "${var.name_prefix}-aurora-connections"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "2"
  metric_name         = "DatabaseConnections"
  namespace           = "AWS/RDS"
  period              = "300"
  statistic           = "Average"
  threshold           = var.max_connections * 0.8
  alarm_description   = "This metric monitors Aurora cluster connections"
  alarm_actions       = var.alarm_actions

  dimensions = {
    DBClusterIdentifier = aws_rds_cluster.main.id
  }

  tags = var.tags
}
