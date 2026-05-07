output "cluster_name" {
  description = "AKS cluster name"
  value       = azurerm_kubernetes_cluster.main.name
}

output "cluster_endpoint" {
  description = "AKS API server endpoint (private)"
  value       = azurerm_kubernetes_cluster.main.kube_config[0].host
}

output "oidc_issuer_url" {
  description = "OIDC issuer URL for Workload Identity — used in federated credential"
  value       = azurerm_kubernetes_cluster.main.oidc_issuer_url
}

output "user_assigned_identity_id" {
  description = "ID of the User Assigned Identity — used to bind Azure roles"
  value       = azurerm_user_assigned_identity.aks.id
}

output "kubernetes_version" {
  description = "Deployed Kubernetes version"
  value       = azurerm_kubernetes_cluster.main.kubernetes_version
}

# Mandatory output for IRSA-like binding (symmetric with aws-eks)
output "iam_policy_arn" {
  description = "Placeholder — Azure uses Role Assignments, not IAM policies"
  value       = azurerm_user_assigned_identity.aks.id
}
