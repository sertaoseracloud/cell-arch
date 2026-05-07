variable "project_name" {
  description = "Project identifier"
  type        = string
}

variable "environment" {
  description = "Deployment environment (dev, hom, prod)"
  type        = string
}

variable "aws_region" {
  description = "AWS region"
  type        = string
}

variable "private_subnet_ids" {
  description = "List of private subnet IDs from aws-vpc module (Spoke VPC)"
  type        = list(string)
}

variable "node_count" {
  description = "Desired and minimum node count for the managed node group"
  type        = number
  default     = 2
}

variable "kubernetes_version" {
  description = "EKS Kubernetes version"
  type        = string
  default     = "1.33"
}

variable "namespace" {
  description = "Kubernetes namespace for the IRSA-bound ServiceAccount"
  type        = string
  default     = "default"
}

variable "service_account_name" {
  description = "Kubernetes ServiceAccount name for IRSA binding"
  type        = string
  default     = "sidecar"
}

variable "tags" {
  description = "Mandatory resource tags"
  type        = map(string)
  default     = {}
}
