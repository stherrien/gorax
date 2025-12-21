# IAM Module Outputs

output "api_service_role_arn" {
  description = "ARN of the API service role"
  value       = aws_iam_role.api_service.arn
}

output "api_service_role_name" {
  description = "Name of the API service role"
  value       = aws_iam_role.api_service.name
}

output "worker_service_role_arn" {
  description = "ARN of the worker service role"
  value       = aws_iam_role.worker_service.arn
}

output "worker_service_role_name" {
  description = "Name of the worker service role"
  value       = aws_iam_role.worker_service.name
}

output "eks_cluster_role_arn" {
  description = "ARN of the EKS cluster role"
  value       = aws_iam_role.eks_cluster.arn
}

output "eks_cluster_role_name" {
  description = "Name of the EKS cluster role"
  value       = aws_iam_role.eks_cluster.name
}

output "eks_node_group_role_arn" {
  description = "ARN of the EKS node group role"
  value       = aws_iam_role.eks_node_group.arn
}

output "eks_node_group_role_name" {
  description = "Name of the EKS node group role"
  value       = aws_iam_role.eks_node_group.name
}

output "cluster_autoscaler_role_arn" {
  description = "ARN of the cluster autoscaler role"
  value       = aws_iam_role.cluster_autoscaler.arn
}

output "cluster_autoscaler_role_name" {
  description = "Name of the cluster autoscaler role"
  value       = aws_iam_role.cluster_autoscaler.name
}

output "load_balancer_controller_role_arn" {
  description = "ARN of the load balancer controller role"
  value       = aws_iam_role.load_balancer_controller.arn
}

output "load_balancer_controller_role_name" {
  description = "Name of the load balancer controller role"
  value       = aws_iam_role.load_balancer_controller.name
}

output "ebs_csi_driver_role_arn" {
  description = "ARN of the EBS CSI driver role"
  value       = aws_iam_role.ebs_csi_driver.arn
}

output "ebs_csi_driver_role_name" {
  description = "Name of the EBS CSI driver role"
  value       = aws_iam_role.ebs_csi_driver.name
}
