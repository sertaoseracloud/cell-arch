output "eks_cluster_name" {
  description = "EKS cluster name"
  value       = module.eks.cluster_name;
}

output "eks_cluster_endpoint" {
  description = "EKS private API endpoint"
  value       = module.eks.cluster_endpoint;
}

output "dynamodb_table_name" {
  description = "DynamoDB table name"
  value       = module.dynamodb.table_name;
}
