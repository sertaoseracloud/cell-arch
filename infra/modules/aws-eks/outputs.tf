output "cluster_name" {
  description = "EKS cluster name"
  value       = aws_eks_cluster.main.name
}

output "cluster_endpoint" {
  description = "Private EKS API server endpoint"
  value       = aws_eks_cluster.main.endpoint
}

output "cluster_ca" {
  description = "Base64-encoded cluster CA certificate"
  value       = aws_eks_cluster.main.certificate_authority[0].data
  sensitive   = true
}

output "oidc_issuer_url" {
  description = "OIDC issuer URL for IRSA"
  value       = aws_eks_cluster.main.identity[0].oidc[0].issuer
}

output "oidc_provider_arn" {
  description = "ARN of the IAM OIDC provider"
  value       = aws_iam_openid_connect_provider.eks.arn
}

output "irsa_role_arn" {
  description = "ARN of the IRSA IAM role"
  value       = aws_iam_role.irsa.arn
}

output "irsa_role_name" {
  description = "Name of the IRSA IAM role"
  value       = aws_iam_role.irsa.name
}

output "node_group_name" {
  description = "Name of the managed node group"
  value       = aws_eks_node_group.main.node_group_name
}

output "iam_policy_arn" {
  description = "Placeholder — DynamoDB policy ARN attached in live root"
  value       = aws_iam_role.irsa.arn
}
