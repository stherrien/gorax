# Terraform Backend Configuration
# Store state in S3 with DynamoDB locking

# Note: This backend configuration should be uncommented and configured after
# creating the S3 bucket and DynamoDB table for state management.
#
# To initialize:
# 1. Create S3 bucket: aws s3 mb s3://gorax-terraform-state-<account-id>
# 2. Enable versioning: aws s3api put-bucket-versioning --bucket gorax-terraform-state-<account-id> --versioning-configuration Status=Enabled
# 3. Create DynamoDB table: aws dynamodb create-table --table-name gorax-terraform-locks --attribute-definitions AttributeName=LockID,AttributeType=S --key-schema AttributeName=LockID,KeyType=HASH --billing-mode PAY_PER_REQUEST
# 4. Uncomment the backend block below and run: terraform init -migrate-state

# terraform {
#   backend "s3" {
#     bucket         = "gorax-terraform-state-<account-id>"
#     key            = "gorax/terraform.tfstate"
#     region         = "us-east-1"
#     encrypt        = true
#     dynamodb_table = "gorax-terraform-locks"
#
#     # Optional: Use KMS for state encryption
#     # kms_key_id = "arn:aws:kms:us-east-1:<account-id>:key/<key-id>"
#   }
# }

# For environment-specific state files, use workspace or separate state keys:
# - Dev:        key = "gorax/dev/terraform.tfstate"
# - Staging:    key = "gorax/staging/terraform.tfstate"
# - Production: key = "gorax/production/terraform.tfstate"
