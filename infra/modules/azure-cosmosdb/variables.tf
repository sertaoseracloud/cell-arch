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

variable "endpoint_subnet_id" {
  description = "Subnet ID for CosmosDB Private Endpoint (from azure-vnet module)"
  type        = string
}

variable "private_dns_zone_id" {
  description = "ID of the private DNS zone for CosmosDB (created in live root)"
  type        = string
}

variable "tags" {
  description = "Mandatory resource tags"
  type        = map(string)
  default     = {}
}
