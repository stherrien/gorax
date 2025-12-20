# Root Terraform Configuration for Gorax
# Composes all infrastructure modules

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Project     = "Gorax"
      Environment = var.environment
      ManagedBy   = "Terraform"
    }
  }
}

locals {
  name_prefix = "${var.project_name}-${var.environment}"

  common_tags = merge(
    var.tags,
    {
      Environment = var.environment
      Project     = var.project_name
    }
  )
}

# KMS Keys (must be created first)
module "kms" {
  source = "./modules/kms"

  name_prefix               = local.name_prefix
  deletion_window_in_days   = var.kms_deletion_window_days
  tags                      = local.common_tags
}

# VPC and Networking
module "vpc" {
  source = "./modules/vpc"

  name_prefix              = local.name_prefix
  vpc_cidr                 = var.vpc_cidr
  az_count                 = var.availability_zone_count
  cluster_name             = local.name_prefix
  aws_region               = var.aws_region
  enable_flow_logs         = var.enable_vpc_flow_logs
  flow_logs_retention_days = var.vpc_flow_logs_retention_days
  tags                     = local.common_tags
}

# EKS Cluster (depends on VPC)
module "eks" {
  source = "./modules/eks"

  cluster_name            = local.name_prefix
  kubernetes_version      = var.kubernetes_version
  vpc_id                  = module.vpc.vpc_id
  private_subnet_ids      = module.vpc.private_subnet_ids
  public_subnet_ids       = module.vpc.public_subnet_ids
  cluster_role_arn        = module.iam.eks_cluster_role_arn
  node_role_arn           = module.iam.eks_node_group_role_arn
  kms_key_arn             = module.kms.eks_key_arn
  enable_public_access    = var.eks_enable_public_access
  public_access_cidrs     = var.eks_public_access_cidrs
  workstation_cidrs       = var.eks_workstation_cidrs
  cluster_log_types       = var.eks_cluster_log_types
  log_retention_days      = var.eks_log_retention_days
  node_groups             = var.eks_node_groups
  ebs_csi_driver_role_arn = module.iam.ebs_csi_driver_role_arn
  tags                    = local.common_tags

  depends_on = [
    module.vpc,
    module.iam
  ]
}

# IAM Roles (depends on EKS OIDC provider)
module "iam" {
  source = "./modules/iam"

  name_prefix       = local.name_prefix
  oidc_provider_arn = module.eks.oidc_provider_arn
  oidc_provider     = module.eks.oidc_provider_url
  namespace         = var.kubernetes_namespace
  secrets_arns      = ["*"]
  sqs_queue_arns    = module.sqs.all_queue_arns
  s3_bucket_arns    = module.s3.all_bucket_arns
  kms_key_arns = [
    module.kms.database_key_arn,
    module.kms.s3_key_arn,
    module.kms.secrets_key_arn,
    module.kms.elasticache_key_arn,
    module.kms.sqs_key_arn
  ]
  tags = local.common_tags

  depends_on = [
    module.eks
  ]
}

# Aurora PostgreSQL
module "aurora" {
  source = "./modules/aurora"

  name_prefix                = local.name_prefix
  vpc_id                     = module.vpc.vpc_id
  subnet_ids                 = module.vpc.database_subnet_ids
  allowed_security_group_ids = [module.eks.node_security_group_id]
  availability_zones         = module.vpc.availability_zones
  database_name              = var.database_name
  master_username            = var.database_master_username
  engine_version             = var.aurora_engine_version
  engine_family              = var.aurora_engine_family
  enable_serverless_v2       = var.aurora_enable_serverless_v2
  serverless_min_capacity    = var.aurora_serverless_min_capacity
  serverless_max_capacity    = var.aurora_serverless_max_capacity
  instance_count             = var.aurora_instance_count
  backup_retention_period    = var.aurora_backup_retention_period
  deletion_protection        = var.aurora_deletion_protection
  kms_key_arn                = module.kms.database_key_arn
  secrets_kms_key_arn        = module.kms.secrets_key_arn
  monitoring_interval        = var.aurora_monitoring_interval
  enable_performance_insights = var.aurora_enable_performance_insights
  create_cloudwatch_alarms   = var.create_cloudwatch_alarms
  alarm_actions              = var.alarm_sns_topic_arns
  tags                       = local.common_tags

  depends_on = [
    module.vpc,
    module.kms
  ]
}

# ElastiCache Redis
module "elasticache" {
  source = "./modules/elasticache"

  name_prefix                = local.name_prefix
  vpc_id                     = module.vpc.vpc_id
  subnet_ids                 = module.vpc.private_subnet_ids
  allowed_security_group_ids = [module.eks.node_security_group_id]
  engine_version             = var.redis_engine_version
  parameter_group_family     = var.redis_parameter_group_family
  node_type                  = var.redis_node_type
  num_cache_nodes            = var.redis_num_cache_nodes
  automatic_failover_enabled = var.redis_automatic_failover_enabled
  multi_az_enabled           = var.redis_multi_az_enabled
  snapshot_retention_limit   = var.redis_snapshot_retention_limit
  transit_encryption_enabled = var.redis_transit_encryption_enabled
  kms_key_arn                = module.kms.elasticache_key_arn
  secrets_kms_key_arn        = module.kms.secrets_key_arn
  create_cloudwatch_alarms   = var.create_cloudwatch_alarms
  alarm_actions              = var.alarm_sns_topic_arns
  tags                       = local.common_tags

  depends_on = [
    module.vpc,
    module.kms
  ]
}

# SQS Queues
module "sqs" {
  source = "./modules/sqs"

  name_prefix              = local.name_prefix
  kms_key_id               = module.kms.sqs_key_id
  allowed_principal_arns   = [
    module.iam.api_service_role_arn,
    module.iam.worker_service_role_arn
  ]
  visibility_timeout_seconds = var.sqs_visibility_timeout_seconds
  high_priority_visibility_timeout_seconds = var.sqs_high_priority_visibility_timeout_seconds
  message_retention_seconds = var.sqs_message_retention_seconds
  max_receive_count        = var.sqs_max_receive_count
  create_fifo_queue        = var.sqs_create_fifo_queue
  create_cloudwatch_alarms = var.create_cloudwatch_alarms
  alarm_actions            = var.alarm_sns_topic_arns
  tags                     = local.common_tags

  depends_on = [
    module.kms,
    module.iam
  ]
}

# S3 Buckets
module "s3" {
  source = "./modules/s3"

  name_prefix = local.name_prefix
  kms_key_arn = module.kms.s3_key_arn
  allowed_principal_arns = [
    module.iam.api_service_role_arn,
    module.iam.worker_service_role_arn
  ]
  artifacts_transition_to_ia_days = var.s3_artifacts_transition_to_ia_days
  artifacts_transition_to_glacier_days = var.s3_artifacts_transition_to_glacier_days
  logs_transition_to_ia_days = var.s3_logs_transition_to_ia_days
  logs_transition_to_glacier_days = var.s3_logs_transition_to_glacier_days
  logs_expiration_days       = var.s3_logs_expiration_days
  enable_access_logging      = var.s3_enable_access_logging
  create_cloudwatch_alarms   = var.create_cloudwatch_alarms
  alarm_actions              = var.alarm_sns_topic_arns
  tags                       = local.common_tags

  depends_on = [
    module.kms,
    module.iam
  ]
}
