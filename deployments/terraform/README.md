# Gorax AWS Infrastructure - Terraform

This directory contains Terraform configurations for deploying the Gorax workflow automation platform on AWS.

## Architecture Overview

The infrastructure consists of:

- **VPC**: Multi-AZ network with public, private, and database subnets
- **EKS**: Managed Kubernetes cluster with autoscaling node groups
- **Aurora PostgreSQL**: Serverless v2 database cluster with automated backups
- **ElastiCache Redis**: Multi-AZ cache cluster with encryption
- **SQS**: Message queues for workflow execution with DLQ support
- **S3**: Encrypted buckets for artifacts and logs with lifecycle policies
- **IAM**: Service roles with IRSA for EKS workloads
- **KMS**: Encryption keys for all services
- **CloudWatch**: Logs and alarms for monitoring

## Prerequisites

### Required Tools

- [Terraform](https://www.terraform.io/downloads.html) >= 1.5.0
- [AWS CLI](https://aws.amazon.com/cli/) >= 2.0
- [kubectl](https://kubernetes.io/docs/tasks/tools/) >= 1.28

### AWS Credentials

Configure AWS credentials:

```bash
aws configure
# OR
export AWS_ACCESS_KEY_ID="your-access-key"
export AWS_SECRET_ACCESS_KEY="your-secret-key"
export AWS_REGION="us-east-1"
```

### Required Permissions

The AWS user/role needs permissions to create:
- VPC and networking resources
- EKS clusters and node groups
- RDS/Aurora clusters
- ElastiCache clusters
- S3 buckets
- SQS queues
- IAM roles and policies
- KMS keys
- CloudWatch logs and alarms

## Initial Setup

### 1. Create Backend Infrastructure

Before using Terraform, create the S3 bucket and DynamoDB table for state management:

```bash
# Set your AWS account ID
ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)

# Create S3 bucket for state
aws s3 mb s3://gorax-terraform-state-${ACCOUNT_ID} --region us-east-1

# Enable versioning
aws s3api put-bucket-versioning \
  --bucket gorax-terraform-state-${ACCOUNT_ID} \
  --versioning-configuration Status=Enabled

# Enable encryption
aws s3api put-bucket-encryption \
  --bucket gorax-terraform-state-${ACCOUNT_ID} \
  --server-side-encryption-configuration '{
    "Rules": [{
      "ApplyServerSideEncryptionByDefault": {
        "SSEAlgorithm": "AES256"
      }
    }]
  }'

# Block public access
aws s3api put-public-access-block \
  --bucket gorax-terraform-state-${ACCOUNT_ID} \
  --public-access-block-configuration \
    "BlockPublicAcls=true,IgnorePublicAcls=true,BlockPublicPolicy=true,RestrictPublicBuckets=true"

# Create DynamoDB table for locking
aws dynamodb create-table \
  --table-name gorax-terraform-locks \
  --attribute-definitions AttributeName=LockID,AttributeType=S \
  --key-schema AttributeName=LockID,KeyType=HASH \
  --billing-mode PAY_PER_REQUEST \
  --region us-east-1
```

### 2. Configure Backend

Update the backend configuration in the environment files:

```bash
# Replace <account-id> with your AWS account ID in:
# - environments/dev/main.tf
# - environments/staging/main.tf
# - environments/production/main.tf

sed -i "s/<account-id>/${ACCOUNT_ID}/g" environments/*/main.tf
```

## Directory Structure

```
deployments/terraform/
├── modules/                    # Reusable Terraform modules
│   ├── vpc/                   # VPC and networking
│   ├── eks/                   # EKS cluster
│   ├── aurora/                # Aurora PostgreSQL
│   ├── elasticache/           # Redis cluster
│   ├── sqs/                   # SQS queues
│   ├── s3/                    # S3 buckets
│   ├── iam/                   # IAM roles and policies
│   └── kms/                   # KMS encryption keys
├── environments/              # Environment-specific configs
│   ├── dev/                   # Development environment
│   ├── staging/               # Staging environment
│   └── production/            # Production environment
├── main.tf                    # Root module composition
├── variables.tf               # Root variables
├── outputs.tf                 # Root outputs
├── versions.tf                # Provider version constraints
├── backend.tf                 # Backend configuration template
└── README.md                  # This file
```

## Deployment

### Development Environment

```bash
cd environments/dev

# Initialize Terraform
terraform init

# Review the plan
terraform plan

# Apply the configuration
terraform apply

# Save outputs for later use
terraform output > outputs.txt
```

### Staging Environment

```bash
cd environments/staging

terraform init
terraform plan
terraform apply
```

### Production Environment

```bash
cd environments/production

terraform init
terraform plan

# Review carefully before applying to production
terraform apply
```

## Environment Configurations

### Development

- **Cost-optimized**: Uses SPOT instances and minimal resources
- **Single AZ**: Reduced redundancy for cost savings
- **Minimal backups**: 1-day retention
- **No alarms**: Monitoring disabled to reduce costs
- **Small instances**: t3.medium for EKS, cache.t4g.micro for Redis

### Staging

- **Production-like**: Mirrors production configuration at smaller scale
- **Multi-AZ**: High availability enabled
- **7-day backups**: Short-term retention
- **Alarms enabled**: Monitoring active
- **Medium instances**: t3.large for EKS, cache.t4g.small for Redis

### Production

- **High availability**: 3 AZs with redundancy
- **Long backups**: 30-day retention
- **Full monitoring**: All alarms enabled
- **Separate node groups**: API and worker workloads isolated
- **Larger instances**: t3.xlarge for API, t3.large for workers

## Post-Deployment Configuration

### 1. Configure kubectl

```bash
# Get the command from Terraform output
terraform output kubectl_config_command

# Run the command (example)
aws eks update-kubeconfig --region us-east-1 --name gorax-dev

# Verify connection
kubectl get nodes
```

### 2. Create Kubernetes Namespace

```bash
kubectl create namespace gorax
```

### 3. Create Service Accounts with IRSA

Create service accounts for API and worker pods:

```yaml
# api-service-account.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: gorax-api
  namespace: gorax
  annotations:
    eks.amazonaws.com/role-arn: <api-service-role-arn>
---
# worker-service-account.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: gorax-worker
  namespace: gorax
  annotations:
    eks.amazonaws.com/role-arn: <worker-service-role-arn>
```

Replace `<api-service-role-arn>` and `<worker-service-role-arn>` with values from Terraform outputs:

```bash
terraform output api_service_role_arn
terraform output worker_service_role_arn
```

Apply the service accounts:

```bash
kubectl apply -f api-service-account.yaml
kubectl apply -f worker-service-account.yaml
```

### 4. Install Additional EKS Add-ons

#### AWS Load Balancer Controller

```bash
# Add Helm repository
helm repo add eks https://aws.github.io/eks-charts
helm repo update

# Install AWS Load Balancer Controller
helm install aws-load-balancer-controller eks/aws-load-balancer-controller \
  -n kube-system \
  --set clusterName=<cluster-name> \
  --set serviceAccount.create=false \
  --set serviceAccount.name=aws-load-balancer-controller \
  --set region=us-east-1 \
  --set vpcId=<vpc-id>
```

#### Cluster Autoscaler

```bash
# Install Cluster Autoscaler
kubectl apply -f https://raw.githubusercontent.com/kubernetes/autoscaler/master/cluster-autoscaler/cloudprovider/aws/examples/cluster-autoscaler-autodiscover.yaml

# Configure for your cluster
kubectl -n kube-system annotate serviceaccount cluster-autoscaler \
  eks.amazonaws.com/role-arn=<cluster-autoscaler-role-arn>

kubectl -n kube-system set image deployment.apps/cluster-autoscaler \
  cluster-autoscaler=registry.k8s.io/autoscaling/cluster-autoscaler:v1.28.0
```

### 5. Retrieve Database Credentials

```bash
# Get the secret ARN from Terraform output
terraform output database_secret_arn

# Retrieve credentials
aws secretsmanager get-secret-value \
  --secret-id <secret-arn> \
  --query SecretString \
  --output text | jq .
```

### 6. Create Kubernetes Secrets

```bash
# Database credentials
kubectl create secret generic gorax-db-credentials \
  --from-literal=host=<database-endpoint> \
  --from-literal=port=5432 \
  --from-literal=database=gorax \
  --from-literal=username=<from-secrets-manager> \
  --from-literal=password=<from-secrets-manager> \
  -n gorax

# Redis credentials (if transit encryption enabled)
kubectl create secret generic gorax-redis-credentials \
  --from-literal=host=<redis-endpoint> \
  --from-literal=port=6379 \
  --from-literal=auth-token=<from-secrets-manager> \
  -n gorax

# SQS configuration
kubectl create configmap gorax-sqs-config \
  --from-literal=main-queue-url=<main-queue-url> \
  --from-literal=high-priority-queue-url=<high-priority-queue-url> \
  -n gorax

# S3 configuration
kubectl create configmap gorax-s3-config \
  --from-literal=artifacts-bucket=<artifacts-bucket-name> \
  --from-literal=logs-bucket=<logs-bucket-name> \
  -n gorax
```

## Cost Optimization

### Development Environment Costs (Estimated)

- **EKS Cluster**: ~$73/month
- **EKS Nodes** (1x t3.medium SPOT): ~$15/month
- **Aurora Serverless v2** (0.5 ACU): ~$40/month
- **ElastiCache** (1x cache.t4g.micro): ~$12/month
- **NAT Gateway**: ~$33/month
- **Data Transfer**: Variable
- **Total**: ~$173/month

### Cost-Saving Tips

1. **Stop dev environment when not in use**: Scale EKS node groups to 0
2. **Use SPOT instances**: Enabled in dev by default
3. **Reduce Aurora capacity**: Lower min ACU to 0.5
4. **Disable unnecessary features**: VPC Flow Logs, Performance Insights in dev
5. **Shorter retention periods**: Reduce backup and log retention

### Stopping Development Environment

```bash
# Scale node groups to 0
aws eks update-nodegroup-config \
  --cluster-name gorax-dev \
  --nodegroup-name gorax-dev-general \
  --scaling-config desiredSize=0,minSize=0,maxSize=3

# Pause Aurora cluster (not available for Serverless v2, consider using provisioned)
```

## Monitoring and Logging

### CloudWatch Logs

Logs are available in CloudWatch:

- **EKS Control Plane**: `/aws/eks/<cluster-name>/cluster`
- **VPC Flow Logs**: `/aws/vpc/<name-prefix>-flow-logs`
- **Redis Logs**: `/aws/elasticache/<name-prefix>/slow-log` and `engine-log`

### CloudWatch Alarms

The following alarms are created (if enabled):

- **Aurora**: CPU utilization, database connections
- **ElastiCache**: CPU, memory, evictions, replication lag
- **SQS**: Message age, queue depth, DLQ depth
- **S3**: Bucket size

### Viewing Alarms

```bash
# List all alarms
aws cloudwatch describe-alarms --alarm-name-prefix gorax-

# Get alarm history
aws cloudwatch describe-alarm-history \
  --alarm-name gorax-<env>-aurora-cpu-1
```

## Maintenance

### Upgrading Kubernetes Version

1. Update `kubernetes_version` in environment configuration
2. Plan and apply Terraform changes
3. Update node groups (done automatically by Terraform)

```bash
terraform plan
terraform apply
```

### Database Backups

Aurora creates automatic backups daily. To create a manual snapshot:

```bash
aws rds create-db-cluster-snapshot \
  --db-cluster-identifier gorax-<env>-aurora-cluster \
  --db-cluster-snapshot-identifier gorax-manual-snapshot-$(date +%Y%m%d)
```

### Restoring from Backup

```bash
aws rds restore-db-cluster-from-snapshot \
  --db-cluster-identifier gorax-restored \
  --snapshot-identifier <snapshot-id> \
  --engine aurora-postgresql
```

## Troubleshooting

### Terraform State Lock Issues

If Terraform state is locked:

```bash
# List locks
aws dynamodb scan --table-name gorax-terraform-locks

# Force unlock (use with caution)
terraform force-unlock <lock-id>
```

### EKS Authentication Issues

```bash
# Update kubeconfig
aws eks update-kubeconfig --name <cluster-name> --region us-east-1

# Verify IAM user
aws sts get-caller-identity

# Check aws-auth ConfigMap
kubectl get configmap aws-auth -n kube-system -o yaml
```

### Database Connection Issues

```bash
# Verify security group allows connection from EKS nodes
aws ec2 describe-security-groups --group-ids <db-security-group-id>

# Test connection from a pod
kubectl run -it --rm debug --image=postgres:15 --restart=Never -- \
  psql -h <db-endpoint> -U <username> -d gorax
```

### Pod Cannot Assume IAM Role

Check service account annotation:

```bash
kubectl describe serviceaccount gorax-api -n gorax
```

Verify OIDC provider trust relationship:

```bash
aws iam get-role --role-name <role-name> --query 'Role.AssumeRolePolicyDocument'
```

## Security Best Practices

1. **Encryption**: All data is encrypted at rest and in transit
2. **Network isolation**: Private subnets for workloads, database in separate subnet
3. **IAM roles**: Use IRSA instead of instance roles
4. **Secrets**: Store credentials in Secrets Manager, not in code
5. **VPC endpoints**: Reduce data transfer costs and improve security
6. **Security groups**: Principle of least privilege
7. **Deletion protection**: Enabled for production database

## Disaster Recovery

### Backup Strategy

- **Aurora**: Automated daily backups with 7-30 day retention
- **Redis**: Daily snapshots with configurable retention
- **S3**: Versioning enabled, lifecycle policies for archival
- **Terraform state**: Versioned in S3

### Recovery Procedures

1. **Database**: Restore from automated backup or snapshot
2. **Configuration**: Recreate from Terraform state
3. **Application data**: Restore from S3 backups

### Cross-Region DR (Optional)

To enable cross-region replication:

1. Create S3 bucket in secondary region
2. Set `enable_cross_region_replication = true`
3. Configure Aurora global database (manual)

## Cleanup

### Destroy Environment

```bash
cd environments/dev

# Preview what will be destroyed
terraform plan -destroy

# Destroy infrastructure
terraform destroy

# Confirm with 'yes' when prompted
```

**Warning**: This will delete all resources including databases and S3 buckets (if `skip_final_snapshot = true`).

### Cleanup Order Issues

If destroy fails due to dependency issues:

1. Delete Kubernetes resources first: `kubectl delete all --all -n gorax`
2. Delete EKS add-ons: LoadBalancers, PersistentVolumes
3. Run `terraform destroy` again

## Module Reference

### VPC Module

Creates VPC with public, private, and database subnets across multiple AZs.

**Inputs**: `vpc_cidr`, `az_count`, `cluster_name`
**Outputs**: `vpc_id`, `subnet_ids`

### EKS Module

Creates managed Kubernetes cluster with node groups and OIDC provider.

**Inputs**: `cluster_name`, `kubernetes_version`, `node_groups`
**Outputs**: `cluster_endpoint`, `oidc_provider_arn`

### Aurora Module

Creates Aurora PostgreSQL Serverless v2 cluster with automated backups.

**Inputs**: `instance_count`, `serverless_min_capacity`, `serverless_max_capacity`
**Outputs**: `cluster_endpoint`, `secret_arn`

### ElastiCache Module

Creates Redis replication group with Multi-AZ support.

**Inputs**: `node_type`, `num_cache_nodes`, `transit_encryption_enabled`
**Outputs**: `primary_endpoint_address`, `secret_arn`

### SQS Module

Creates message queues with dead letter queues.

**Inputs**: `visibility_timeout_seconds`, `max_receive_count`
**Outputs**: `main_queue_url`, `main_queue_arn`

### S3 Module

Creates buckets for artifacts and logs with lifecycle policies.

**Inputs**: `transition_to_ia_days`, `transition_to_glacier_days`
**Outputs**: `artifacts_bucket_arn`, `logs_bucket_arn`

### IAM Module

Creates service roles for EKS workloads using IRSA.

**Inputs**: `oidc_provider_arn`, `namespace`
**Outputs**: `api_service_role_arn`, `worker_service_role_arn`

### KMS Module

Creates encryption keys for all services.

**Inputs**: `deletion_window_in_days`
**Outputs**: `database_key_arn`, `secrets_key_arn`, `s3_key_arn`

## Support

For issues or questions:

1. Check the troubleshooting section above
2. Review CloudWatch logs for errors
3. Consult AWS documentation for service-specific issues
4. Open an issue in the Gorax repository

## License

This infrastructure code is part of the Gorax project. See the main repository for license details.
