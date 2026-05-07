variable "project_name" {
  description = "Project identifier"
  type        = string
}

variable "environment" {
  description = "Deployment environment (dev, hom, prod)"
  type        = string
}

variable "location" {
  description = "Azure region"
  type        = string
}

variable "private_subnet_ids" {
  description = "List of private subnet IDs from azure-vnet module (Spoke VNet)"
  type        = list(string)
}

variable "node_count" {
  description = "Desired node count for the default node pool"
  type        = number
  default     = 2
}

variable "kubernetes_version" {
  description = "AKS Kubernetes version — verify with `az aks get-versions` before apply"
  type        = string
  default     = "1.29"
}

variable "service_account_name" {
  description = "Kubernetes ServiceAccount name for Workload Identity binding"
  type        = string
  default     = "sidecar"
}

variable "tags" {
  description = "Mandatory resource tags"
  type        = map(string)
  default     = {}
}
