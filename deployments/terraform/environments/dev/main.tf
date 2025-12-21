# Development Environment Configuration

terraform {
  backend "s3" {
    bucket         = "gorax-terraform-state-<account-id>"
    key            = "gorax/dev/terraform.tfstate"
    region         = "us-east-1"
    encrypt        = true
    dynamodb_table = "gorax-terraform-locks"
  }
}

module "gorax" {
  source = "../.."

  # General Configuration
  project_name = "gorax"
  environment  = "dev"
  aws_region   = "us-east-1"

  # VPC Configuration
  vpc_cidr                     = "10.0.0.0/16"
  availability_zone_count      = 2
  enable_vpc_flow_logs         = true
  vpc_flow_logs_retention_days = 7

  # EKS Configuration
  kubernetes_version       = "1.28"
  kubernetes_namespace     = "gorax"
  eks_enable_public_access = true
  eks_public_access_cidrs  = ["0.0.0.0/0"]
  eks_log_retention_days   = 7

  eks_node_groups = {
    general = {
      instance_types  = ["t3.medium"]
      capacity_type   = "SPOT"
      disk_size       = 30
      desired_size    = 1
      max_size        = 3
      min_size        = 1
      max_unavailable = 1
      labels          = {
        workload = "general"
      }
      taints          = []
      tags            = {
        NodeGroup = "general"
      }
    }
  }

  # Aurora Configuration
  database_name                   = "gorax_dev"
  database_master_username        = "gorax_admin"
  aurora_engine_version           = "15.4"
  aurora_enable_serverless_v2     = true
  aurora_serverless_min_capacity  = 0.5
  aurora_serverless_max_capacity  = 1
  aurora_instance_count           = 1
  aurora_backup_retention_period  = 1
  aurora_deletion_protection      = false
  aurora_monitoring_interval      = 0
  aurora_enable_performance_insights = false

  # Redis Configuration
  redis_engine_version             = "7.0"
  redis_node_type                  = "cache.t4g.micro"
  redis_num_cache_nodes            = 1
  redis_automatic_failover_enabled = false
  redis_multi_az_enabled           = false
  redis_snapshot_retention_limit   = 1
  redis_transit_encryption_enabled = false

  # SQS Configuration
  sqs_visibility_timeout_seconds                = 300
  sqs_high_priority_visibility_timeout_seconds = 60
  sqs_message_retention_seconds                 = 86400
  sqs_max_receive_count                         = 3
  sqs_create_fifo_queue                         = false

  # S3 Configuration
  s3_artifacts_transition_to_ia_days      = 30
  s3_artifacts_transition_to_glacier_days = 60
  s3_logs_transition_to_ia_days           = 7
  s3_logs_transition_to_glacier_days      = 30
  s3_logs_expiration_days                 = 90
  s3_enable_access_logging                = false

  # Monitoring Configuration
  create_cloudwatch_alarms = false
  alarm_sns_topic_arns     = []

  # KMS Configuration
  kms_deletion_window_days = 7

  tags = {
    Environment = "dev"
    ManagedBy   = "Terraform"
    CostCenter  = "Engineering"
  }
}
