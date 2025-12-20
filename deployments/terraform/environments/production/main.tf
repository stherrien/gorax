# Production Environment Configuration

terraform {
  backend "s3" {
    bucket         = "gorax-terraform-state-<account-id>"
    key            = "gorax/production/terraform.tfstate"
    region         = "us-east-1"
    encrypt        = true
    dynamodb_table = "gorax-terraform-locks"
  }
}

module "gorax" {
  source = "../.."

  # General Configuration
  project_name = "gorax"
  environment  = "production"
  aws_region   = "us-east-1"

  # VPC Configuration
  vpc_cidr                     = "10.2.0.0/16"
  availability_zone_count      = 3
  enable_vpc_flow_logs         = true
  vpc_flow_logs_retention_days = 30

  # EKS Configuration
  kubernetes_version       = "1.28"
  kubernetes_namespace     = "gorax"
  eks_enable_public_access = true
  eks_public_access_cidrs  = ["0.0.0.0/0"]
  eks_log_retention_days   = 30

  eks_node_groups = {
    general = {
      instance_types  = ["t3.xlarge"]
      capacity_type   = "ON_DEMAND"
      disk_size       = 50
      desired_size    = 3
      max_size        = 10
      min_size        = 2
      max_unavailable = 1
      labels          = {
        workload = "general"
      }
      taints          = []
      tags            = {
        NodeGroup = "general"
      }
    }
    workers = {
      instance_types  = ["t3.large"]
      capacity_type   = "ON_DEMAND"
      disk_size       = 50
      desired_size    = 3
      max_size        = 15
      min_size        = 2
      max_unavailable = 1
      labels          = {
        workload = "workers"
      }
      taints          = []
      tags            = {
        NodeGroup = "workers"
      }
    }
  }

  # Aurora Configuration
  database_name                   = "gorax_production"
  database_master_username        = "gorax_admin"
  aurora_engine_version           = "15.4"
  aurora_enable_serverless_v2     = true
  aurora_serverless_min_capacity  = 1
  aurora_serverless_max_capacity  = 4
  aurora_instance_count           = 3
  aurora_backup_retention_period  = 30
  aurora_deletion_protection      = true
  aurora_monitoring_interval      = 60
  aurora_enable_performance_insights = true

  # Redis Configuration
  redis_engine_version             = "7.0"
  redis_node_type                  = "cache.t4g.medium"
  redis_num_cache_nodes            = 3
  redis_automatic_failover_enabled = true
  redis_multi_az_enabled           = true
  redis_snapshot_retention_limit   = 30
  redis_transit_encryption_enabled = true

  # SQS Configuration
  sqs_visibility_timeout_seconds                = 300
  sqs_high_priority_visibility_timeout_seconds = 60
  sqs_message_retention_seconds                 = 345600
  sqs_max_receive_count                         = 3
  sqs_create_fifo_queue                         = false

  # S3 Configuration
  s3_artifacts_transition_to_ia_days      = 90
  s3_artifacts_transition_to_glacier_days = 180
  s3_logs_transition_to_ia_days           = 30
  s3_logs_transition_to_glacier_days      = 90
  s3_logs_expiration_days                 = 365
  s3_enable_access_logging                = true

  # Monitoring Configuration
  create_cloudwatch_alarms = true
  alarm_sns_topic_arns = [
    # Add your SNS topic ARNs here for production alerts
    # "arn:aws:sns:us-east-1:<account-id>:gorax-production-alerts"
  ]

  # KMS Configuration
  kms_deletion_window_days = 30

  tags = {
    Environment = "production"
    ManagedBy   = "Terraform"
    CostCenter  = "Production"
    Criticality = "High"
  }
}
